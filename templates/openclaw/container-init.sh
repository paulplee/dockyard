#!/bin/bash
# =============================================================================
# dockyard / openclaw — Container initialization
# Runs as root at each boot before ssh.service via systemd oneshot unit.
# Handles tasks that need volume mounts to be in place first.
# =============================================================================
set -e

AGENT_USER="${AGENT_USER:-agent}"

# Regenerate SSH host keys if missing
ssh-keygen -A

# Ensure Neovim directories exist and are owned by the agent user
mkdir -p /home/${AGENT_USER}/.local/share/nvim
mkdir -p /home/${AGENT_USER}/.local/state/nvim
chown -R ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.local

# Fix authorized_keys permissions (bind-mounted from host)
if [ -f /home/${AGENT_USER}/.ssh/authorized_keys ]; then
    chmod 600 /home/${AGENT_USER}/.ssh/authorized_keys
    chown ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.ssh/authorized_keys
fi

# Tighten OpenClaw state directory + config permissions (doctor warnings)
if [ -d /home/${AGENT_USER}/.openclaw ]; then
    chmod 700 /home/${AGENT_USER}/.openclaw
    [ -f /home/${AGENT_USER}/.openclaw/openclaw.json ] && chmod 600 /home/${AGENT_USER}/.openclaw/openclaw.json
    [ -d /home/${AGENT_USER}/.openclaw/credentials ] && chmod 700 /home/${AGENT_USER}/.openclaw/credentials
fi
