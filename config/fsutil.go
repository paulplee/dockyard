package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

// sudo returns an *exec.Cmd that runs `sudo <args...>` with stdin, stdout, and
// stderr connected to the terminal so that password prompts work.
func sudo(args ...string) *exec.Cmd {
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// MkdirAllPrivileged creates dir (and parents). If direct creation fails with
// EACCES — typically because $VolumesRoot is root-owned — it retries via sudo.
func MkdirAllPrivileged(dir string, perm os.FileMode) error {
	if err := os.MkdirAll(dir, perm); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrPermission) {
		return err
	}
	return sudo("mkdir", "-p", "-m", fmt.Sprintf("%o", perm), dir).Run()
}

// WriteFilePrivileged writes data to path. Falls back to `sudo tee` on EACCES.
func WriteFilePrivileged(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		if err := os.WriteFile(path, data, perm); err == nil {
			return nil
		} else if !errors.Is(err, fs.ErrPermission) {
			return err
		}
	}
	// sudo tee needs data on stdin, not the terminal — but sudo itself may
	// need the terminal for a password prompt. Use -S to read the password
	// from stderr/terminal while piping data on stdin.
	cmd := exec.Command("sudo", "tee", path)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return sudo("chmod", fmt.Sprintf("%o", perm), path).Run()
}

// RemoveAllPrivileged removes a directory tree, retrying under sudo on EACCES.
func RemoveAllPrivileged(path string) error {
	if err := os.RemoveAll(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrPermission) {
		return err
	}
	return sudo("rm", "-rf", path).Run()
}

// ChownPrivileged sets ownership, always via sudo (chown is privileged).
func ChownPrivileged(path string, uid, gid int, recursive bool) error {
	args := []string{"chown"}
	if recursive {
		args = append(args, "-R")
	}
	args = append(args, fmt.Sprintf("%d:%d", uid, gid), path)
	return sudo(args...).Run()
}

// ChmodPrivileged sets mode, via sudo when necessary.
func ChmodPrivileged(path string, mode os.FileMode, recursive bool) error {
	if err := os.Chmod(path, mode); err == nil && !recursive {
		return nil
	}
	args := []string{"chmod"}
	if recursive {
		args = append(args, "-R")
	}
	args = append(args, fmt.Sprintf("%o", mode), path)
	return sudo(args...).Run()
}

// ChownToSelf chowns path to the calling user's UID:GID via sudo. This is
// used after MkdirAllPrivileged to make a sudo-created directory writable by
// the deploying user.
func ChownToSelf(path string) error {
	uid := os.Getuid()
	if uid == 0 {
		return nil // already root
	}
	return ChownPrivileged(path, uid, os.Getgid(), false)
}

// groupExists reports whether the named group is present on the host.
func groupExists(name string) bool {
	if runtime.GOOS == "darwin" {
		return exec.Command("dscl", ".", "-read", "/Groups/"+name).Run() == nil
	}
	return exec.Command("getent", "group", name).Run() == nil
}

// EnsureGroup creates the named host group (if it does not exist) and adds the
// current user. Equivalent to the Make-era `group` target.
func EnsureGroup(name string, gid int) error {
	fmt.Printf(">>> Ensuring host group '%s' (GID=%d)...\n", name, gid)

	if !groupExists(name) {
		if runtime.GOOS == "darwin" {
			if err := sudo("dscl", ".", "-create", "/Groups/"+name).Run(); err != nil {
				return fmt.Errorf("create group %s: %w", name, err)
			}
			if err := sudo("dscl", ".", "-create", "/Groups/"+name, "PrimaryGroupID", fmt.Sprintf("%d", gid)).Run(); err != nil {
				return fmt.Errorf("set GID for group %s: %w", name, err)
			}
		} else {
			if err := sudo("groupadd", "-g", fmt.Sprintf("%d", gid), name).Run(); err != nil {
				return fmt.Errorf("create group %s: %w", name, err)
			}
		}
	}
	fmt.Printf("  group '%s' ready\n", name)

	// Add current user to the group.
	u, err := user.Current()
	if err != nil {
		return err
	}
	if runtime.GOOS == "darwin" {
		if err := sudo("dseditgroup", "-o", "edit", "-a", u.Username, "-t", "user", name).Run(); err != nil {
			return fmt.Errorf("add %s to group %s: %w", u.Username, name, err)
		}
	} else {
		if err := sudo("usermod", "-aG", name, u.Username).Run(); err != nil {
			return fmt.Errorf("add %s to group %s: %w", u.Username, name, err)
		}
	}
	fmt.Printf("  added %s to group '%s'\n", u.Username, name)
	fmt.Println("  NOTE: re-login (or start a new shell) for group membership to take effect")
	return nil
}
