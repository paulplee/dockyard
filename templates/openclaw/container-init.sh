#!/bin/bash
# =============================================================================
# dockyard / openclaw — Container initialization
# Runs as root at each boot before ssh.service via systemd oneshot unit.
# Handles tasks that need volume mounts to be in place first.
# =============================================================================
set -e

AGENT_USER="${AGENT_USER:-dy-user}"

# Regenerate SSH host keys if missing
ssh-keygen -A

# Seed LazyVim, TPM, and fix Neovim data dirs
source /boot-seed.sh

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

# Install openclaw-gateway systemd user service into bind-mounted ~/.config
SVC_DIR="/home/${AGENT_USER}/.config/systemd/user"
if [ ! -f "${SVC_DIR}/openclaw-gateway.service" ]; then
    echo ">>> Installing openclaw-gateway systemd user service..."
    su -s /bin/bash "${AGENT_USER}" -c \
        "mkdir -p ${SVC_DIR}/default.target.wants \
         && cp /usr/local/share/dockyard/openclaw-gateway.service ${SVC_DIR}/openclaw-gateway.service \
         && ln -sf ../openclaw-gateway.service ${SVC_DIR}/default.target.wants/openclaw-gateway.service"
fi
