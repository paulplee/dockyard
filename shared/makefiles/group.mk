# =============================================================================
# dockyard — shared Makefile: group management
# Included by each template's Makefile.
# Provides: make group
# =============================================================================

AGENT_GROUP  ?= agents
DEPLOY_USER  := $(shell whoami)

.PHONY: group

# Create the shared 'agents' group on the host and add the deploying user.
# Run once per host. Log out and back in (or 'newgrp agents') for membership.
group:
	@echo ">>> Ensuring host group '$(AGENT_GROUP)' (GID=$(AGENT_GID)) exists..."
	@getent group $(AGENT_GROUP) >/dev/null 2>&1 \
		|| sudo groupadd -g $(AGENT_GID) $(AGENT_GROUP) \
		&& echo "  group '$(AGENT_GROUP)' ready"
	@echo ">>> Adding $(DEPLOY_USER) to group '$(AGENT_GROUP)'..."
	@sudo usermod -aG $(AGENT_GROUP) $(DEPLOY_USER)
	@echo "  NOTE: run 'newgrp $(AGENT_GROUP)' or re-login for membership to take effect"
