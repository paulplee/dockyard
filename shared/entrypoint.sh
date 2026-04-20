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

# Fix authorized_keys permissions every start
if [ -f /home/${AGENT_USER}/.ssh/authorized_keys ]; then
  chmod 600 /home/${AGENT_USER}/.ssh/authorized_keys
  chown ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.ssh/authorized_keys
fi

/usr/sbin/sshd

exec "$@"
