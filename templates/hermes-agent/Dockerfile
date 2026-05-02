# =============================================================================
# dockyard / hermes-agent — Nous Research Hermes Agent
# Extends dev-env with Python-based Hermes Agent framework.
# Base    : debian:trixie-slim  (Debian 13)
# =============================================================================
FROM debian:trixie-slim

ARG AGENT_USER=agent
ARG AGENT_UID=1100
ARG NEOVIM_VERSION=0.12.1
ARG SSH_PORT=22
# Changing this value busts the Docker layer cache for the venv build step,
# forcing pip to re-fetch hermes-agent from git on the next `dockyard deploy`.
ARG HERMES_VERSION=main

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8 \
    TZ=Asia/Hong_Kong

# ── Locale & timezone ───────────────────────────────────────────────────────
RUN apt-get update && apt-get install -y --no-install-recommends \
    locales tzdata \
    && sed -i '/en_US.UTF-8/s/^# //' /etc/locale.gen \
    && locale-gen \
    && ln -sf /usr/share/zoneinfo/Asia/Hong_Kong /etc/localtime \
    && dpkg-reconfigure -f noninteractive tzdata \
    && rm -rf /var/lib/apt/lists/*

# ── System packages ─────────────────────────────────────────────────────────
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash curl wget git unzip tar gzip \
    openssh-server \
    build-essential cmake ninja-build pkg-config \
    python3 python3-venv python3-dev \
    python3-pip \
    nodejs npm \
    bat \
    ripgrep fd-find fzf jq htop btop procps tree \
    ca-certificates dnsutils iputils-ping \
    tmux \
    sudo \
    systemd systemd-sysv dbus libpam-systemd \
    && rm -rf /var/lib/apt/lists/* \
    && systemctl enable ssh dbus \
    && for unit in \
    getty@.service \
    systemd-udevd.service \
    systemd-udevd-kernel.socket \
    systemd-udevd-control.socket; do \
    ln -sf /dev/null /etc/systemd/system/"${unit}"; \
    done

# 'bat' is installed as 'batcat' on Debian — alias it
RUN ln -sf /usr/bin/batcat /usr/local/bin/bat

# ── Neovim (latest stable) ──────────────────────────────────────────────────
# TARGETARCH is set by Docker BuildKit ("amd64"/"arm64"); uname -m is the
# fallback for native builds without buildx ("x86_64"/"aarch64").
ARG TARGETARCH
RUN case "${TARGETARCH:-$(uname -m)}" in \
    amd64|x86_64)  NVIM_ARCH=nvim-linux-x86_64 ;; \
    arm64|aarch64) NVIM_ARCH=nvim-linux-arm64  ;; \
    *) echo "Unsupported arch: ${TARGETARCH:-$(uname -m)}" && exit 1 ;; \
    esac \
    && curl -fsSL \
    "https://github.com/neovim/neovim/releases/download/v${NEOVIM_VERSION}/${NVIM_ARCH}.tar.gz" \
    -o /tmp/nvim.tar.gz \
    && tar -C /opt -xzf /tmp/nvim.tar.gz \
    && ln -sf /opt/${NVIM_ARCH}/bin/nvim /usr/local/bin/nvim \
    && rm /tmp/nvim.tar.gz

# ── Agent OS user ────────────────────────────────────────────────────────────
RUN groupadd -g ${AGENT_UID} ${AGENT_USER} \
    && useradd -m -u ${AGENT_UID} -g ${AGENT_USER} -s /bin/bash ${AGENT_USER} \
    && echo "${AGENT_USER} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/${AGENT_USER} \
    && chmod 0440 /etc/sudoers.d/${AGENT_USER}

# ── User-space setup (LazyVim + tmux) ────────────────────────────────────────
USER ${AGENT_USER}

# LazyVim bootstrap
RUN git clone --depth 1 \
    https://github.com/LazyVim/starter \
    /home/${AGENT_USER}/.config/nvim \
    && mkdir -p \
    /home/${AGENT_USER}/.local/share/nvim \
    /home/${AGENT_USER}/.local/state/nvim \
    /home/${AGENT_USER}/.cache/nvim

# Pre-warm plugins so the first SSH login is instant
RUN nvim --headless "+Lazy! sync" +qa 2>/dev/null || true

# tmux config and TPM
RUN mkdir -p /home/${AGENT_USER}/.config/tmux/plugins \
    && git clone https://github.com/paulplee/mancave /tmp/mancave \
    && cp /tmp/mancave/config_files/.tmux.conf /home/${AGENT_USER}/.config/tmux/tmux.conf \
    && rm -rf /tmp/mancave \
    && git clone https://github.com/tmux-plugins/tpm /home/${AGENT_USER}/.config/tmux/plugins/tpm

# Ensure XDG_RUNTIME_DIR is always set — needed for systemctl --user in
# docker exec sessions that bypass PAM (libpam-systemd only sets it for
# PAM-authenticated logins such as SSH; docker exec skips PAM entirely).
RUN echo 'export XDG_RUNTIME_DIR=/run/user/$(id -u)' \
    >> /home/${AGENT_USER}/.bashrc

USER root

# ── Python virtual environment with Hermes Agent ─────────────────────────────
# Venv lives in the agent user's home so it can be bind-mounted as a volume
# and survive container rebuilds. A seed copy is kept in /opt/hermes-venv-seed
# so container-init.sh can populate an empty volume on first boot.
RUN python3 -m venv /home/${AGENT_USER}/.hermes-venv \
    && /home/${AGENT_USER}/.hermes-venv/bin/pip install --no-cache-dir \
    "hermes-agent[web] @ git+https://github.com/NousResearch/hermes-agent.git@${HERMES_VERSION}" \
    openai \
    anthropic \
    litellm \
    httpx \
    pydantic \
    structlog \
    rich \
    tenacity \
    python-telegram-bot \
    && SITE_PKG=$(/home/${AGENT_USER}/.hermes-venv/bin/python3 -c "import site; print(site.getsitepackages()[0])") \
    && git clone --depth 1 --branch "${HERMES_VERSION}" https://github.com/NousResearch/hermes-agent.git /tmp/hermes-agent-web-src \
    && npm --prefix /tmp/hermes-agent-web-src/web install \
    && npm --prefix /tmp/hermes-agent-web-src/web run build \
    && cp -r /tmp/hermes-agent-web-src/hermes_cli/web_dist "${SITE_PKG}/hermes_cli/" \
    && rm -rf /tmp/hermes-agent-web-src \
    && chown -R ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.hermes-venv \
    && cp -a /home/${AGENT_USER}/.hermes-venv /opt/hermes-venv-seed \
    && date -u +%Y%m%dT%H%M%SZ > /opt/hermes-venv-seed/.hermes-venv-buildtime \
    && rm -rf /root/.cache /root/.npm

# ── SSH hardening ────────────────────────────────────────────────────────────
RUN mkdir -p /var/run/sshd \
    && sed -i \
    -e 's/#PasswordAuthentication yes/PasswordAuthentication no/' \
    -e 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' \
    -e 's/#PermitRootLogin prohibit-password/PermitRootLogin no/' \
    -e "s/#Port 22/Port ${SSH_PORT}/" \
    /etc/ssh/sshd_config \
    && echo "AllowUsers ${AGENT_USER}" >> /etc/ssh/sshd_config \
    && echo "StrictModes no" >> /etc/ssh/sshd_config

# authorized_keys is injected via bind mount at runtime
RUN mkdir -p /home/${AGENT_USER}/.ssh \
    && chown ${AGENT_USER}:${AGENT_USER} /home/${AGENT_USER}/.ssh \
    && chmod 700 /home/${AGENT_USER}/.ssh

# ── Workspace layout ────────────────────────────────────────────────────────
# /workspace  — persistent code & data (bind-mounted from host)
# /secrets    — read-only API keys / tokens
# /logs       — structured log output
RUN mkdir -p /workspace /secrets /logs \
    && chown ${AGENT_USER}:${AGENT_USER} /workspace /logs \
    && chmod 750 /secrets

# ── Hermes Agent systemd service ─────────────────────────────────────────────
# Installed to a stable image path; container-init.sh links it into the
# bind-mounted ~/.config/systemd/user/ on each boot.
COPY hermes-agent.service /usr/local/share/dockyard/hermes-agent.service
COPY hermes-dashboard.service /usr/local/share/dockyard/hermes-dashboard.service

# Enable linger so user services start on boot without an interactive login
RUN mkdir -p /var/lib/systemd/linger \
    && touch /var/lib/systemd/linger/${AGENT_USER}

# ── Container init script + boot-seed ────────────────────────────────────────
COPY container-init.sh /usr/local/sbin/container-init.sh
COPY container-init.service /etc/systemd/system/container-init.service
COPY boot-seed.sh /boot-seed.sh
RUN chmod 755 /usr/local/sbin/container-init.sh /boot-seed.sh \
    && mkdir -p /etc/systemd/system/multi-user.target.wants \
    && ln -sf /etc/systemd/system/container-init.service \
    /etc/systemd/system/multi-user.target.wants/container-init.service

ENV AGENT_USER=${AGENT_USER} \
    HERMES_VENV=/home/${AGENT_USER}/.hermes-venv \
    PATH=/home/${AGENT_USER}/.hermes-venv/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

EXPOSE ${SSH_PORT}

STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["/lib/systemd/systemd"]
