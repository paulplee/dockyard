#!/bin/bash
# =============================================================================
# dockyard / hermes-agent — Container initialization
# Runs as root at each boot before ssh.service via systemd oneshot unit.
# Handles tasks that need volume mounts to be in place first.
# =============================================================================
set -e

AGENT_USER="${AGENT_USER:-dy-user}"

# Regenerate SSH host keys if missing
ssh-keygen -A

# Seed base config and fix volume permissions
source /boot-seed.sh

# Fix authorized_keys permissions (bind-mounted from host)
if [ -f /home/${AGENT_USER}/.ssh/authorized_keys ]; then
    chmod 600 /home/${AGENT_USER}/.ssh/authorized_keys
    chown ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.ssh/authorized_keys
fi

# Tighten Hermes Agent data directory permissions
if [ -d /home/${AGENT_USER}/.hermes-agent ]; then
    chmod 700 /home/${AGENT_USER}/.hermes-agent
    # Secure any config files
    find /home/${AGENT_USER}/.hermes-agent -name "*.json" -exec chmod 600 {} \; 2>/dev/null || true
    find /home/${AGENT_USER}/.hermes-agent -name "*.yaml" -exec chmod 600 {} \; 2>/dev/null || true
    find /home/${AGENT_USER}/.hermes-agent -name "*.yml" -exec chmod 600 {} \; 2>/dev/null || true
fi

# Seed hermes venv from image backup on first boot (volume starts empty)
if [ ! -f /home/${AGENT_USER}/.hermes-venv/bin/python ]; then
    echo ">>> Seeding Hermes venv from image seed..."
    cp -a /opt/hermes-venv-seed/. /home/${AGENT_USER}/.hermes-venv/
    chown -R ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.hermes-venv
fi

# Install hermes-agent systemd user service into bind-mounted ~/.config
SVC_DIR="/home/${AGENT_USER}/.config/systemd/user"
if [ ! -f "${SVC_DIR}/hermes-agent.service" ]; then
    echo ">>> Installing hermes-agent systemd user service..."
    su -s /bin/bash "${AGENT_USER}" -c \
        "mkdir -p ${SVC_DIR}/default.target.wants \
         && cp /usr/local/share/dockyard/hermes-agent.service ${SVC_DIR}/hermes-agent.service \
         && ln -sf ../hermes-agent.service ${SVC_DIR}/default.target.wants/hermes-agent.service"
fi

# Ensure the user systemd instance is running so the D-Bus session socket
# exists when hermes-agent calls systemctl --user.
AGENT_UID="$(id -u "${AGENT_USER}")"
if ! systemctl is-active --quiet "user@${AGENT_UID}.service" 2>/dev/null; then
    echo ">>> Starting user@${AGENT_UID} systemd instance..."
    systemctl start "user@${AGENT_UID}.service" || true
fi
