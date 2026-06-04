# =============================================================================
# dockyard / private — Closed-source private application scaffold
# Base    : debian:trixie-slim  (Debian 13)
# =============================================================================
FROM debian:trixie-slim

ARG AGENT_USER=agent
ARG AGENT_UID=1100
ARG NEOVIM_VERSION=0.12.1
ARG GLOW_VERSION=2.1.0
ARG SSH_PORT=22
# Changing this value busts the Docker layer cache for the private app
# installation step, forcing a rebuild on the next `dockyard deploy`.
ARG PRIVATE_APP_VERSION=main
ARG PRIVATE_APP_PORT=7700

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
    less \
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

# ── Glow (markdown renderer) ────────────────────────────────────────────────
RUN case "${TARGETARCH:-$(uname -m)}" in \
    amd64|x86_64)  GLOW_ARCH=x86_64 ;; \
    arm64|aarch64) GLOW_ARCH=arm64  ;; \
    *) echo "Unsupported arch: ${TARGETARCH:-$(uname -m)}" && exit 1 ;; \
    esac \
    && curl -fsSL \
    "https://github.com/charmbracelet/glow/releases/download/v${GLOW_VERSION}/glow_${GLOW_VERSION}_Linux_${GLOW_ARCH}.tar.gz" \
    -o /tmp/glow.tar.gz \
    && mkdir -p /tmp/glow-extract \
    && tar -C /tmp/glow-extract -xzf /tmp/glow.tar.gz \
    && find /tmp/glow-extract -name 'glow' -type f -exec mv {} /usr/local/bin/glow \; \
    && chmod 755 /usr/local/bin/glow \
    && rm -rf /tmp/glow.tar.gz /tmp/glow-extract

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

# ── Private app installation ──────────────────────────────────────────────────
# Uncomment ONE of the patterns below and customise for your application.
# This section is intentionally left commented — the template ships no code.
#
# Pattern A — Build from private git repository (requires BuildKit secret)
# RUN --mount=type=secret,id=private_repo_token \
#     TOKEN=$(cat /run/secrets/private_repo_token) \
#     git clone https://oauth2:${TOKEN}@github.com/<org>/<repo>.git \
#           --branch ${PRIVATE_APP_VERSION} --depth 1 /opt/private-app \
#     && cd /opt/private-app \
#     && <build steps> \
#     && ln -sf /opt/private-app/<binary> /usr/local/bin/private-app
#
# Pattern B1 — Download prebuilt binary from private URL (via BuildKit secret)
# RUN --mount=type=secret,id=private_repo_token \
#     TOKEN=$(cat /run/secrets/private_repo_token) \
#     && curl -fsSL -H "Authorization: Bearer ${TOKEN}" \
#        "https://releases.example.com/private-app-${PRIVATE_APP_VERSION}-linux-amd64" \
#        -o /usr/local/bin/private-app \
#     && chmod +x /usr/local/bin/private-app
#
# Pattern B2 — COPY binary placed next to Dockerfile
# COPY private-app /usr/local/bin/private-app
# RUN chmod +x /usr/local/bin/private-app
#
# Note: If your app needs a systemd user service, place the .service file
# in the image and wire it up in container-init.sh (see the commented block
# at the end of container-init.sh).

# ── Workspace layout ────────────────────────────────────────────────────────
# /workspace  — persistent code & data (bind-mounted from host)
# /secrets    — read-only API keys / tokens
# /logs       — structured log output
RUN mkdir -p /workspace /secrets /logs \
    && chown ${AGENT_USER}:${AGENT_USER} /workspace /logs \
    && chmod 750 /secrets

# ── Container init script + boot-seed + entrypoint ──────────────────────────
COPY container-init.sh /usr/local/sbin/container-init.sh
COPY container-init.service /etc/systemd/system/container-init.service
COPY entrypoint.sh /entrypoint.sh
COPY boot-seed.sh /boot-seed.sh
RUN chmod 755 /usr/local/sbin/container-init.sh /entrypoint.sh /boot-seed.sh \
    && mkdir -p /etc/systemd/system/multi-user.target.wants \
    && ln -sf /etc/systemd/system/container-init.service \
    /etc/systemd/system/multi-user.target.wants/container-init.service

ENV AGENT_USER=${AGENT_USER}

EXPOSE ${SSH_PORT}

STOPSIGNAL SIGRTMIN+3
CMD ["/entrypoint.sh", "/sbin/init"]
