# =============================================================================
# dockyard / dev-env — SSH-accessible development environment
# Base    : debian:trixie-slim  (Debian 13)
# Tools   : Neovim (LazyVim), tmux (TPM), Python 3, Node.js, ripgrep, bat, …
# =============================================================================
FROM debian:trixie-slim

ARG AGENT_USER=agent
ARG AGENT_UID=1100
ARG NEOVIM_VERSION=0.12.1
ARG SSH_PORT=22

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
    python3 python3-pip python3-venv python3-dev \
    nodejs npm \
    bat \
    ripgrep fd-find fzf jq htop btop procps tree \
    ca-certificates dnsutils iputils-ping \
    tmux \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# 'bat' is installed as 'batcat' on Debian — alias it
RUN ln -sf /usr/bin/batcat /usr/local/bin/bat

# ── uv (Python package manager) ─────────────────────────────────────────────
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /usr/local/bin/

# ── Neovim (latest stable) ──────────────────────────────────────────────────
# Resolve the Neovim release tarball name for the current build architecture.
# Docker sets TARGETARCH to "amd64" or "arm64"; map that to the nvim asset name.
RUN case "$(uname -m)" in \
    x86_64)  NVIM_ARCH=nvim-linux-x86_64 ;; \
    aarch64) NVIM_ARCH=nvim-linux-arm64  ;; \
    *) echo "Unsupported arch: $(uname -m)" && exit 1 ;; \
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

# ── Workspace layout ────────────────────────────────────────────────────────
# /workspace  — persistent code & data (bind-mounted from host)
# /secrets    — read-only API keys / tokens
# /logs       — structured log output
RUN mkdir -p /workspace /secrets /logs \
    && chown ${AGENT_USER}:${AGENT_USER} /workspace /logs \
    && chmod 750 /secrets

# ── Entrypoint ───────────────────────────────────────────────────────────────
COPY entrypoint.sh /entrypoint.sh
COPY boot-seed.sh /boot-seed.sh
RUN chmod +x /entrypoint.sh /boot-seed.sh

ENV AGENT_USER=${AGENT_USER}

EXPOSE ${SSH_PORT}

ENTRYPOINT ["/entrypoint.sh"]
