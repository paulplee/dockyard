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

# Seed hermes venv from image backup on first boot or when the image has been
# rebuilt (detected via a build-time stamp written into the seed during image
# build). This ensures that dependency updates (e.g. new extras like [web])
# propagate to existing bind-mounted volumes automatically.
SEED_BUILDTIME="/opt/hermes-venv-seed/.hermes-venv-buildtime"
VENV_BUILDTIME="/home/${AGENT_USER}/.hermes-venv/.hermes-venv-buildtime"
if [ ! -f "${VENV_BUILDTIME}" ] || ! diff -q "${SEED_BUILDTIME}" "${VENV_BUILDTIME}" >/dev/null 2>&1; then
    echo ">>> Seeding Hermes venv from image seed..."
    # /home/${AGENT_USER}/.hermes-venv is a bind-mount point — we cannot
    # remove the directory itself, only its contents.
    find /home/${AGENT_USER}/.hermes-venv -mindepth 1 -delete 2>/dev/null || true
    cp -a /opt/hermes-venv-seed/. /home/${AGENT_USER}/.hermes-venv/
    chown -R ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.hermes-venv
fi

# Install hermes-agent and hermes-dashboard systemd user services into bind-mounted ~/.config
SVC_DIR="/home/${AGENT_USER}/.config/systemd/user"
if [ ! -f "${SVC_DIR}/hermes-agent.service" ]; then
    echo ">>> Installing hermes-agent systemd user service..."
    su -s /bin/bash "${AGENT_USER}" -c \
        "mkdir -p ${SVC_DIR}/default.target.wants \
         && cp /usr/local/share/dockyard/hermes-agent.service ${SVC_DIR}/hermes-agent.service \
         && ln -sf ../hermes-agent.service ${SVC_DIR}/default.target.wants/hermes-agent.service"
fi
if [ ! -f "${SVC_DIR}/hermes-dashboard.service" ]; then
    echo ">>> Installing hermes-dashboard systemd user service..."
    su -s /bin/bash "${AGENT_USER}" -c \
        "mkdir -p ${SVC_DIR}/default.target.wants \
         && cp /usr/local/share/dockyard/hermes-dashboard.service ${SVC_DIR}/hermes-dashboard.service \
         && ln -sf ../hermes-dashboard.service ${SVC_DIR}/default.target.wants/hermes-dashboard.service"
fi

AGENT_UID="$(id -u "${AGENT_USER}")"

# Ensure the user systemd instance is running so the D-Bus session socket
# exists when we start user services below.
if ! systemctl is-active --quiet "user@${AGENT_UID}.service" 2>/dev/null; then
    echo ">>> Starting user@${AGENT_UID} systemd instance..."
    systemctl start "user@${AGENT_UID}.service" || true
fi

# Start user services now — the user@UID instance may have started before the
# services were installed by this script, so kick them explicitly.
DBUS="unix:path=/run/user/${AGENT_UID}/bus"
su -s /bin/bash "${AGENT_USER}" -c \
    "XDG_RUNTIME_DIR=/run/user/${AGENT_UID} DBUS_SESSION_BUS_ADDRESS=${DBUS} \
     systemctl --user start hermes-agent.service hermes-dashboard.service 2>/dev/null || true"
