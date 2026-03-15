# Turnpike - Build System
# Usage: make [target]

APP_NAME    := turnpike
VERSION     := 2.0.3
BUILD_DIR   := build
CMD_PATH    := ./cmd/turnpike
LDFLAGS     := -s -w -X 'github.com/KilimcininKorOglu/Turnpike/internal/cli.AppVersion=$(VERSION)'
GO          := go
GOTEST      := $(GO) test
GOBUILD     := $(GO) build

# Platform targets
PLATFORMS := \
	windows/amd64 \
	windows/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64

.PHONY: all build test clean run version help
.PHONY: build-windows build-darwin build-linux
.PHONY: build-all lint vet fmt tidy

# ─────────────────────────────────────────────────
# Default target: build for all platforms
# ─────────────────────────────────────────────────

all: clean build-all

# ─────────────────────────────────────────────────
# Build targets
# ─────────────────────────────────────────────────

CURRENT_OS   := $(shell go env GOOS)
CURRENT_ARCH := $(shell go env GOARCH)
CURRENT_EXT  := $(if $(filter windows,$(CURRENT_OS)),.exe,)

build: ## Build for current platform
	@echo "Building $(APP_NAME) for $(CURRENT_OS)/$(CURRENT_ARCH)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-$(CURRENT_OS)-$(CURRENT_ARCH)$(CURRENT_EXT) $(CMD_PATH)
	@echo "Output: $(BUILD_DIR)/$(APP_NAME)-$(CURRENT_OS)-$(CURRENT_ARCH)$(CURRENT_EXT)"

build-all: $(PLATFORMS) ## Build for all platforms
	@echo ""
	@echo "All builds complete:"
	@ls -lh $(BUILD_DIR)/$(APP_NAME)-* 2>/dev/null || true

$(PLATFORMS):
	$(eval GOOS := $(word 1,$(subst /, ,$@)))
	$(eval GOARCH := $(word 2,$(subst /, ,$@)))
	$(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
	$(eval OUTPUT := $(BUILD_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH)$(EXT))
	@echo "Building $(OUTPUT)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(OUTPUT) $(CMD_PATH) 2>/dev/null || \
		(CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(OUTPUT) $(CMD_PATH) 2>/dev/null || \
		echo "  Skipped $(GOOS)/$(GOARCH) (cross-compilation toolchain not available)")

build-windows: ## Build for Windows (amd64 + arm64)
	@$(MAKE) --no-print-directory windows/amd64 windows/arm64

build-darwin: ## Build for macOS (amd64 + arm64)
	@$(MAKE) --no-print-directory darwin/amd64 darwin/arm64

build-linux: ## Build for Linux (amd64 + arm64)
	@$(MAKE) --no-print-directory linux/amd64 linux/arm64

# ─────────────────────────────────────────────────
# Test targets
# ─────────────────────────────────────────────────

test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) ./internal/... -count=1

test-verbose: ## Run all tests with verbose output
	$(GOTEST) ./internal/... -count=1 -v

test-race: ## Run tests with race detector
	$(GOTEST) ./internal/... -count=1 -race

test-cover: ## Run tests with coverage report
	@mkdir -p $(BUILD_DIR)
	$(GOTEST) ./internal/... -count=1 -coverprofile=$(BUILD_DIR)/coverage.out
	$(GO) tool cover -func=$(BUILD_DIR)/coverage.out
	@echo ""
	@echo "HTML report: $(BUILD_DIR)/coverage.html"
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

test-pkg: ## Run tests for a specific package (usage: make test-pkg PKG=auth)
	$(GOTEST) ./internal/$(PKG)/... -count=1 -v

# ─────────────────────────────────────────────────
# Code quality
# ─────────────────────────────────────────────────

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Format code
	$(GO) fmt ./...

tidy: ## Tidy module dependencies
	$(GO) mod tidy

lint: vet fmt ## Run all linters (vet + fmt)

# ─────────────────────────────────────────────────
# Run & utility
# ─────────────────────────────────────────────────

run: ## Run the application (GUI mode)
	$(GO) run -ldflags "$(LDFLAGS)" $(CMD_PATH)

run-cli: ## Run CLI version check
	$(GO) run -ldflags "$(LDFLAGS)" $(CMD_PATH) --version

clean: ## Remove build artifacts
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."

version: ## Show version
	@echo "$(APP_NAME) v$(VERSION)"

# ─────────────────────────────────────────────────
# Help
# ─────────────────────────────────────────────────

help: ## Show this help
	@echo "$(APP_NAME) v$(VERSION) - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
