# =============================================================================
# dockyard — top-level Makefile
# =============================================================================

.PHONY: list help

help: list

list:
	@echo ""
	@echo "  Available templates:"
	@echo "  ────────────────────"
	@for dir in templates/*/; do \
		name=$$(basename "$$dir"); \
		echo "    $$name"; \
	done
	@echo ""
	@echo "  Usage:"
	@echo "    cd templates/<name>"
	@echo "    make setup"
	@echo "    make deploy AGENT_NAME=<instance>"
	@echo ""
