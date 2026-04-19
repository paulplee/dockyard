# =============================================================================
# dockyard — shared Makefile: Docker Compose operations
# Included by each template's Makefile.
# Provides: down, logs, shell, clean, reset, deploy, guard-container
# Note:     'up' is defined by each template (to allow prepare + --build)
# Expects:  COMPOSE, CONTAINER_NAME, VOLUMES_BASE, VOLUMES_ROOT, AGENT_DIRS
# =============================================================================

COMPOSE ?= docker compose

.PHONY: up down deploy guard-container ssh-config logs shell clean reset

guard-container:
	@if [ -z "$(CONTAINER_NAME)" ]; then \
	  echo ""; \
	  echo "  ERROR: CONTAINER_NAME is not set."; \
	  echo ""; \
	  echo "  Known deployments in $(VOLUMES_ROOT):"; \
	  if [ -d "$(VOLUMES_ROOT)" ]; then \
	    found=$$(ls -1 "$(VOLUMES_ROOT)" 2>/dev/null); \
	    if [ -n "$$found" ]; then \
	      echo "$$found" | sed 's/^/    /'; \
	    else \
	      echo "    (none yet — run 'make setup' first)"; \
	    fi; \
	  else \
	    echo "    ($(VOLUMES_ROOT) does not exist — run 'make setup' first)"; \
	  fi; \
	  echo ""; \
	  echo "  Usage:  make <target> CONTAINER_NAME=<name>"; \
	  echo "          make <target> c=<name>"; \
	  echo ""; \
	  exit 1; \
	fi
	@if [ ! -f "$(VOLUMES_BASE)/.env" ]; then \
	  echo ""; \
	  echo "  ERROR: Unknown deployment '$(CONTAINER_NAME)' — no .env found at $(VOLUMES_BASE)/.env"; \
	  echo ""; \
	  echo "  Known deployments in $(VOLUMES_ROOT):"; \
	  if [ -d "$(VOLUMES_ROOT)" ]; then \
	    found=$$(ls -1 "$(VOLUMES_ROOT)" 2>/dev/null | while read d; do \
	      [ -f "$(VOLUMES_ROOT)/$$d/.env" ] && echo "$$d"; done); \
	    if [ -n "$$found" ]; then \
	      echo "$$found" | sed 's/^/    /'; \
	    else \
	      echo "    (none yet — run 'make setup' first)"; \
	    fi; \
	  fi; \
	  echo ""; \
	  echo "  Run 'make setup' to register a new deployment."; \
	  echo ""; \
	  exit 1; \
	fi
	@if [ "$(filter reset,$(MAKECMDGOALS))" = "" ]; then \
	  running=$$(docker ps --format '{{.Names}}' 2>/dev/null | grep "^dockyard-$(CONTAINER_NAME)$$" || true); \
	  if [ -n "$$running" ]; then \
	    echo ""; \
	    echo "  WARNING: Container 'dockyard-$(CONTAINER_NAME)' is already running."; \
	    echo "  To force a fresh deploy:  make reset c=$(CONTAINER_NAME)"; \
	    echo "  To stop and remove only:  make clean c=$(CONTAINER_NAME)"; \
	    echo ""; \
	    exit 1; \
	  fi; \
	fi

down:
	$(COMPOSE) --env-file $(VOLUMES_BASE)/.env down

deploy: guard-container group init up

logs:
	docker logs -f dy-$(CONTAINER_NAME)

shell:
	docker exec -it dy-$(CONTAINER_NAME) bash

# Remove containers, images, networks, and host volume data (preserves ssh keys)
clean:
	@echo ">>> Stopping and removing containers, networks, images..."
	$(COMPOSE) --env-file $(VOLUMES_BASE)/.env down --rmi all --remove-orphans || true
	@echo ">>> Removing host volume data (preserving ssh keys and secrets)..."
	@sudo rm -rf \
		$(VOLUMES_BASE)/nvim-data \
		$(VOLUMES_BASE)/nvim-state \
		$(VOLUMES_BASE)/workspace \
		$(VOLUMES_BASE)/logs
	@echo ">>> Clean complete. SSH keys preserved in $(VOLUMES_BASE)/ssh/"

# Full teardown + fresh deploy
reset: clean deploy

# Add (or verify) the ~/.ssh/config entry for an existing deployment
ssh-config:
	@if [ -z "$(CONTAINER_NAME)" ]; then \
	  echo "ERROR: set c=<name>"; exit 1; \
	fi
	@if [ ! -f "$(VOLUMES_BASE)/.env" ]; then \
	  echo "ERROR: no .env found at $(VOLUMES_BASE)/.env — run 'make setup c=$(CONTAINER_NAME)' first"; \
	  exit 1; \
	fi
	@sshcfg="$(HOME)/.ssh/config"; \
	 stanza="# dockyard: dy-$(CONTAINER_NAME)"; \
	 port="$(SSH_PORT)"; \
	 keyfile=$$(ls $(HOME)/.ssh/*.pub 2>/dev/null | head -1 | sed 's/\.pub$$//'); \
	 [ -z "$$keyfile" ] && keyfile="$(HOME)/.ssh/id_ed25519"; \
	 if ! grep -qF "$$stanza" "$$sshcfg" 2>/dev/null; then \
	   printf "\n$$stanza\nHost dy-$(CONTAINER_NAME)\n    HostName 127.0.0.1\n    Port $$port\n    User agent\n    IdentityFile $$keyfile\n    IdentitiesOnly yes\n    StrictHostKeyChecking accept-new\n    UserKnownHostsFile $(HOME)/.config/dockyard/known_hosts\n" >> "$$sshcfg"; \
	   chmod 600 "$$sshcfg"; \
	   echo "Added 'dy-$(CONTAINER_NAME)' to ~/.ssh/config  →  ssh dy-$(CONTAINER_NAME)"; \
	 else \
	   echo "SSH config entry 'dy-$(CONTAINER_NAME)' already exists"; \
	 fi
