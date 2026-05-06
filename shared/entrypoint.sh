#!/usr/bin/env bash
# =============================================================================
# dockyard — shared entrypoint
# Starts sshd in the background, seeds first-boot config, then exec's CMD.
# =============================================================================
set -euo pipefail

AGENT_USER="${AGENT_USER:-dy-user}"

# Regenerate SSH host keys on first boot
ssh-keygen -A

# Seed LazyVim, TPM, and fix Neovim data dirs
source /boot-seed.sh

# Fix .ssh directory ownership and permissions.
# The ssh/ host dir is root-owned (root_dir); SSH client requires agent-owned 700.
if [ -d /home/${AGENT_USER}/.ssh ]; then
  chown ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.ssh
  chmod 700 /home/${AGENT_USER}/.ssh
fi

# Fix authorized_keys permissions (inside the bind-mounted .ssh/ directory)
if [ -f /home/${AGENT_USER}/.ssh/authorized_keys ]; then
  chmod 600 /home/${AGENT_USER}/.ssh/authorized_keys
  chown ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.ssh/authorized_keys
fi

# Ensure a persistent ed25519 identity key exists for git operations.
# If the user pre-placed id_ed25519 + id_ed25519.pub in $VOLUMES_BASE/ssh/ this is skipped.
SSH_KEY="/home/${AGENT_USER}/.ssh/id_ed25519"
if [ ! -f "${SSH_KEY}" ]; then
  echo ">>> Generating git SSH identity key..."
  ssh-keygen -t ed25519 -f "${SSH_KEY}" -N "" \
      -C "${CONTAINER_NAME:-$(hostname)}"
  chown ${AGENT_USER}:${AGENT_USER} "${SSH_KEY}" "${SSH_KEY}.pub"
  chmod 600 "${SSH_KEY}"
  chmod 644 "${SSH_KEY}.pub"
  cp "${SSH_KEY}.pub" /logs/git-ssh-pubkey.txt
  echo ">>> Git SSH public key written to /logs/git-ssh-pubkey.txt — add it to GitHub/GitLab."
fi

# Write ~/.ssh/config stanzas for GitHub and GitLab if not already present
SSH_CFG="/home/${AGENT_USER}/.ssh/config"
if ! grep -q "Host github.com" "${SSH_CFG}" 2>/dev/null; then
  cat >> "${SSH_CFG}" <<'SSHCFG'

Host github.com
  IdentityFile ~/.ssh/id_ed25519
  StrictHostKeyChecking accept-new

Host gitlab.com
  IdentityFile ~/.ssh/id_ed25519
  StrictHostKeyChecking accept-new
SSHCFG
  chown ${AGENT_USER}:${AGENT_USER} "${SSH_CFG}"
  chmod 600 "${SSH_CFG}"
fi

# Configure git user identity from /secrets/env (GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL).
# Writes to ~/.config/git/config which persists via the config/ volume mount.
if [ -f /secrets/env ]; then
  GIT_NAME="$(grep -E '^GIT_AUTHOR_NAME=' /secrets/env | head -1 | sed 's/^GIT_AUTHOR_NAME=//')"
  GIT_EMAIL="$(grep -E '^GIT_AUTHOR_EMAIL=' /secrets/env | head -1 | sed 's/^GIT_AUTHOR_EMAIL=//')"
  if [ -n "${GIT_NAME}" ] || [ -n "${GIT_EMAIL}" ]; then
    GIT_CFG_DIR="/home/${AGENT_USER}/.config/git"
    mkdir -p "${GIT_CFG_DIR}"
    chown ${AGENT_USER}:${AGENT_USER} "${GIT_CFG_DIR}"
    GIT_CFG="${GIT_CFG_DIR}/config"
    [ -n "${GIT_NAME}" ]  && git config -f "${GIT_CFG}" user.name  "${GIT_NAME}"
    [ -n "${GIT_EMAIL}" ] && git config -f "${GIT_CFG}" user.email "${GIT_EMAIL}"
    chown ${AGENT_USER}:${AGENT_USER} "${GIT_CFG}"
    chmod 600 "${GIT_CFG}"
  fi
fi

/usr/sbin/sshd

exec "$@"
