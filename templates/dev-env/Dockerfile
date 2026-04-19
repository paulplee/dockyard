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
    ripgrep fd-find fzf jq htop btop procps \
    ca-certificates dnsutils iputils-ping \
    tmux \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# 'bat' is installed as 'batcat' on Debian — alias it
RUN ln -sf /usr/bin/batcat /usr/local/bin/bat

# ── uv (Python package manager) ─────────────────────────────────────────────
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /usr/local/bin/

# ── Neovim (latest stable) ──────────────────────────────────────────────────
RUN curl -fsSL \
    "https://github.com/neovim/neovim/releases/download/v${NEOVIM_VERSION}/nvim-linux-x86_64.tar.gz" \
    -o /tmp/nvim.tar.gz \
    && tar -C /opt -xzf /tmp/nvim.tar.gz \
    && ln -sf /opt/nvim-linux-x86_64/bin/nvim /usr/local/bin/nvim \
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

USER root

# ── SSH hardening ────────────────────────────────────────────────────────────
RUN mkdir -p /var/run/sshd \
    && sed -i \
        -e 's/#PasswordAuthentication yes/PasswordAuthentication no/' \
        -e 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' \
        -e 's/#PermitRootLogin prohibit-password/PermitRootLogin no/' \
        -e "s/#Port 22/Port ${SSH_PORT}/" \
        /etc/ssh/sshd_config \
    && echo "AllowUsers ${AGENT_USER}" >> /etc/ssh/sshd_config

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
RUN chmod +x /entrypoint.sh

ENV AGENT_USER=${AGENT_USER}

EXPOSE ${SSH_PORT}

ENTRYPOINT ["/entrypoint.sh"]
