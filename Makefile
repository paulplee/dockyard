# =============================================================================
# dockyard — top-level Makefile
# =============================================================================

.PHONY: list help go-build go-install clean-go

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
	@echo "  Usage (Make):"
	@echo "    cd templates/<name>"
	@echo "    make setup"
	@echo "    make deploy CONTAINER_NAME=<instance>"
	@echo ""
	@echo "  Usage (Go CLI — experimental):"
	@echo "    make go-build            # build ./bin/dockyard"
	@echo "    ./bin/dockyard init      # first-time host setup"
	@echo "    ./bin/dockyard create openclaw <name>"
	@echo "    ./bin/dockyard deploy <name>"
	@echo ""

# Build the Go CLI into ./bin/dockyard.
go-build:
	@mkdir -p bin
	go build -o bin/dockyard ./cmd/dockyard
	@echo "  built ./bin/dockyard"

# Install into $GOBIN / ~/go/bin.
go-install:
	go install ./cmd/dockyard
	@echo "  installed $$(go env GOBIN)/dockyard (or ~/go/bin/dockyard)"

clean-go:
	rm -rf bin/
