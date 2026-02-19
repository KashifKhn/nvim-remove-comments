.DEFAULT_GOAL := help

CLI_DIR := cli
BIN_NAME := remove-comments
BIN_OUT  := $(CLI_DIR)/$(BIN_NAME)

# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help
help:
	@printf "Usage: make <target>\n\n"
	@printf "Plugin (Lua)\n"
	@printf "  lint-lua        Run luacheck on lua/ and plugin/\n"
	@printf "  fmt-lua         Format lua/ and plugin/ with stylua\n"
	@printf "  fmt-lua-check   Check lua formatting without writing\n\n"
	@printf "CLI (Go)\n"
	@printf "  build           Build the CLI binary to cli/remove-comments\n"
	@printf "  run             Run the CLI in dry-run mode against the repo root\n"
	@printf "  test            Run all Go tests\n"
	@printf "  test-pkg        Run tests for a single package  make test-pkg PKG=./internal/parser/...\n"
	@printf "  test-run        Run a single test by name       make test-run NAME=TestRemoveComments\n"
	@printf "  lint            Run golangci-lint on cli/\n"
	@printf "  fmt             Format Go source with goimports (fallback: gofmt)\n"
	@printf "  fmt-check       Check Go formatting without writing\n"
	@printf "  vet             Run go vet\n"
	@printf "  tidy            Run go mod tidy\n"
	@printf "  clean           Remove the built binary\n\n"
	@printf "Combined\n"
	@printf "  check           Run all linters and tests (plugin + CLI)\n"

# ── Plugin (Lua) ──────────────────────────────────────────────────────────────

.PHONY: lint-lua
lint-lua:
	luacheck lua/ plugin/

.PHONY: fmt-lua
fmt-lua:
	stylua lua/ plugin/

.PHONY: fmt-lua-check
fmt-lua-check:
	stylua --check lua/ plugin/

# ── CLI (Go) ──────────────────────────────────────────────────────────────────

.PHONY: build
build:
	cd $(CLI_DIR) && go build -o $(BIN_NAME) .

.PHONY: run
run: build
	./$(BIN_OUT) .

.PHONY: test
test:
	cd $(CLI_DIR) && go test ./...

.PHONY: test-pkg
test-pkg:
	cd $(CLI_DIR) && go test $(PKG)

.PHONY: test-run
test-run:
	cd $(CLI_DIR) && go test -run $(NAME) ./...

.PHONY: test-verbose
test-verbose:
	cd $(CLI_DIR) && go test -v ./...

.PHONY: lint
lint:
	cd $(CLI_DIR) && golangci-lint run ./...

.PHONY: fmt
fmt:
	@if command -v goimports > /dev/null 2>&1; then \
		cd $(CLI_DIR) && goimports -w .; \
	else \
		cd $(CLI_DIR) && gofmt -w .; \
	fi

.PHONY: fmt-check
fmt-check:
	@if command -v goimports > /dev/null 2>&1; then \
		cd $(CLI_DIR) && goimports -l . | grep . && exit 1 || exit 0; \
	else \
		cd $(CLI_DIR) && gofmt -l . | grep . && exit 1 || exit 0; \
	fi

.PHONY: vet
vet:
	cd $(CLI_DIR) && go vet ./...

.PHONY: tidy
tidy:
	cd $(CLI_DIR) && go mod tidy

.PHONY: clean
clean:
	rm -f $(BIN_OUT)

# ── Combined ──────────────────────────────────────────────────────────────────

.PHONY: check
check: lint-lua fmt-lua-check vet lint test
