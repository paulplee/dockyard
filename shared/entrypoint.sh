#!/usr/bin/env bash
# =============================================================================
# dockyard — shared entrypoint
# Starts sshd in the background, fixes volume permissions, then exec's CMD.
# =============================================================================
set -euo pipefail

AGENT_USER="${AGENT_USER:-agent}"

# Regenerate SSH host keys on first boot
ssh-keygen -A

# Ensure Neovim data directories exist and are owned by the agent user
mkdir -p /home/${AGENT_USER}/.local/share/nvim
mkdir -p /home/${AGENT_USER}/.local/state/nvim
chown -R ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.local

# Seed LazyVim starter config on first boot
if [ ! -f /home/${AGENT_USER}/.config/nvim/init.lua ]; then
  echo ">>> Seeding LazyVim starter config..."
  su -s /bin/bash "${AGENT_USER}" -c \
    "git clone --depth 1 https://github.com/LazyVim/starter /home/${AGENT_USER}/.config/nvim 2>/dev/null || true"
fi

# Seed TPM on first boot
if [ ! -d /home/${AGENT_USER}/.config/tmux/plugins/tpm ]; then
  echo ">>> Seeding Tmux Plugin Manager..."
  su -s /bin/bash "${AGENT_USER}" -c \
    "mkdir -p /home/${AGENT_USER}/.config/tmux/plugins && git clone --depth 1 https://github.com/tmux-plugins/tpm /home/${AGENT_USER}/.config/tmux/plugins/tpm 2>/dev/null || true"
fi

# Fix authorized_keys permissions every start
if [ -f /home/${AGENT_USER}/.ssh/authorized_keys ]; then
  chmod 600 /home/${AGENT_USER}/.ssh/authorized_keys
  chown ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.ssh/authorized_keys
fi

/usr/sbin/sshd

exec "$@"
