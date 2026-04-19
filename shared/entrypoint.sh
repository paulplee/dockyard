#!/usr/bin/env bash
# =============================================================================
# dockyard — shared entrypoint
# Starts sshd in the background, fixes volume permissions, then exec's CMD.
# =============================================================================
set -euo pipefail

AGENT_USER="${AGENT_USER:-agent}"

# Regenerate SSH host keys on first boot
ssh-keygen -A

# Ensure Neovim parent directories exist and are owned by the agent user
mkdir -p /home/${AGENT_USER}/.local/share/nvim
mkdir -p /home/${AGENT_USER}/.local/state/nvim
chown -R ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.local
chmod -R 700 /home/${AGENT_USER}/.local || true

# Fix authorized_keys permissions every start
if [ -f /home/${AGENT_USER}/.ssh/authorized_keys ]; then
  chmod 600 /home/${AGENT_USER}/.ssh/authorized_keys
  chown ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.ssh/authorized_keys
fi

/usr/sbin/sshd

exec "$@"
