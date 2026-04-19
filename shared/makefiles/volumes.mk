# =============================================================================
# dockyard — shared Makefile: host volume management
# Included by each template's Makefile.
# Provides: make init
# Expects:  AGENT_DIRS, ROOT_DIRS, AGENT_UID, AGENT_GID, VOLUMES_BASE
# =============================================================================

.PHONY: init

init:
	@echo ">>> Creating host volume directories..."
	@sudo mkdir -p $(AGENT_DIRS) $(ROOT_DIRS)
	@for d in $(AGENT_DIRS) $(ROOT_DIRS); do echo "  OK  $$d"; done
	@echo ">>> Setting ownership (UID=$(AGENT_UID), GID=$(AGENT_GID))..."
	@sudo chown -R $(AGENT_UID):$(AGENT_GID) $(AGENT_DIRS)
	@sudo chmod -R 2770 $(AGENT_DIRS)
	@echo ">>> Setting secrets permissions (root-owned, 750)..."
	@sudo chmod 750 $(VOLUMES_BASE)/secrets
	@echo ">>> Init complete."
