GO ?= go

.PHONY: check-go
check-go:
	@command -v $(GO) >/dev/null 2>&1 || { echo "Go is required. Install from https://go.dev/dl/"; exit 1; }

.PHONY: install
install: check-go
	$(GO) install ./cmd/agentshot
	@bin_dir="$$( $(GO) env GOBIN )"; \
	if [ -z "$$bin_dir" ]; then bin_dir="$$( $(GO) env GOPATH )/bin"; fi; \
	if ! echo ":$$PATH:" | grep -q ":$$bin_dir:"; then \
		printf '\033[1;33mNote:\033[0m %s is not in PATH.\n' "$$bin_dir"; \
		printf '\033[1;33mRun:\033[0m echo "export PATH=\\"%s:$$PATH\\"" >> ~/.zprofile && source ~/.zprofile\n' "$$bin_dir"; \
		printf '      (bash: ~/.bash_profile, fallback: ~/.profile)\n'; \
	fi

.PHONY: check
check: check-go
	$(GO) test ./...
	$(GO) vet ./...

.PHONY: uninstall
uninstall:
	@bin_dir="$$( $(GO) env GOBIN 2>/dev/null )"; \
	if [ -z "$$bin_dir" ]; then \
		if [ -n "$$GOBIN" ]; then bin_dir="$$GOBIN"; \
		elif [ -n "$$GOPATH" ]; then bin_dir="$$GOPATH/bin"; \
		else bin_dir="$$HOME/go/bin"; fi; \
	fi; \
	bin_path="$$bin_dir/agentshot"; \
	if [ -f "$$bin_path" ]; then rm -f "$$bin_path"; echo "Removed $$bin_path"; else echo "agentshot not found in $$bin_dir"; fi
