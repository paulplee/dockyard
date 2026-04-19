# =============================================================================
# dockyard — shared Makefile: Docker Compose operations
# Included by each template's Makefile.
# Provides: up, down, logs, shell, clean, reset, deploy
# Expects:  COMPOSE, CONTAINER_NAME, VOLUMES_BASE, AGENT_DIRS
# =============================================================================

COMPOSE ?= docker compose

.PHONY: up down deploy logs shell clean reset

up:
	$(COMPOSE) --env-file $(VOLUMES_BASE)/.env up -d

down:
	$(COMPOSE) --env-file $(VOLUMES_BASE)/.env down

deploy: group init up

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
