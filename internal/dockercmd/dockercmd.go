// Package dockercmd wraps invocations of `docker` and `docker compose`.
package dockercmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Compose runs a docker compose subcommand from the given project directory
// with --env-file envPath. Stdout/stderr stream to the caller's terminal.
func Compose(projectDir, envPath string, extraEnv []string, args ...string) error {
	full := append([]string{"compose", "--env-file", envPath}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Dir = projectDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd.Run()
}

// ContainerState returns the docker-reported state ("running", "exited",
// "restarting", etc.) for the given container name. Returns "absent" if the
// container does not exist.
func ContainerState(containerName string) (string, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", containerName)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		if strings.Contains(errb.String(), "No such object") || strings.Contains(errb.String(), "no such container") {
			return "absent", nil
		}
		return "", fmt.Errorf("docker inspect: %w: %s", err, errb.String())
	}
	return strings.TrimSpace(out.String()), nil
}

// Exec runs `docker exec` interactively against the given container.
func Exec(containerName string, args ...string) error {
	full := append([]string{"exec", "-it", containerName}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Logs tails `docker logs` for the given container.
func Logs(containerName string, follow bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerName)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
