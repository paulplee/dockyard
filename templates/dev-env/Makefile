# =============================================================================
# dockyard / dev-env — Makefile
# =============================================================================

# ── Required: pass AGENT_NAME on the command line or export it ───────────────
# Example: make deploy AGENT_NAME=openclaw-paul
ifndef AGENT_NAME
  $(warning AGENT_NAME is not set. Run: make setup  or  make <target> AGENT_NAME=<name>)
endif

# ── Paths ────────────────────────────────────────────────────────────────────
VOLUMES_ROOT := /tank/docker/volumes/agents
VOLUMES_BASE := $(VOLUMES_ROOT)/$(AGENT_NAME)
SHARED_DIR   := ../../shared

# Load .env from the deployment volume path (outside the repo)
-include $(VOLUMES_BASE)/.env
export

AGENT_GID ?= $(AGENT_UID)

# Directories owned by the agent user
AGENT_DIRS := \
	$(VOLUMES_BASE)/nvim-data \
	$(VOLUMES_BASE)/nvim-state \
	$(VOLUMES_BASE)/workspace \
	$(VOLUMES_BASE)/logs

# Directories that stay root-owned
ROOT_DIRS := \
	$(VOLUMES_BASE)/ssh \
	$(VOLUMES_BASE)/secrets

# Export VOLUMES_BASE so docker-compose.yml can use it
export VOLUMES_BASE

# ── Copy shared entrypoint into build context before building ────────────────
.PHONY: prepare

prepare:
	@cp $(SHARED_DIR)/entrypoint.sh ./entrypoint.sh

# Override 'up' dependency to include prepare
up: prepare
	docker compose --env-file $(VOLUMES_BASE)/.env up -d --build

deploy: group init up

# ── Include shared targets ───────────────────────────────────────────────────
include $(SHARED_DIR)/makefiles/setup.mk
include $(SHARED_DIR)/makefiles/group.mk
include $(SHARED_DIR)/makefiles/volumes.mk
include $(SHARED_DIR)/makefiles/docker.mk
