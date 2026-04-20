#!/usr/bin/env bash
# =============================================================================
# dockyard — shared first-boot seeding
# Seeds LazyVim and TPM on first container start. Expects $AGENT_USER to be set.
# Designed to be sourced from entrypoint.sh or container-init.sh.
# =============================================================================

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
