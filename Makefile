.PHONY: help install build test test-coverage lint format clean publish release examples check-version bump-patch bump-minor bump-major dev

SHELL := /bin/bash

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOVET := $(GOCMD) vet
GOLINT := golangci-lint

# Package info
PACKAGE_NAME := github.com/tavor-dev/sdk-go
VERSION_FILE := version.go
CURRENT_VERSION := $(shell grep -oE '[0-9]+\.[0-9]+\.[0-9]+' $(VERSION_FILE) 2>/dev/null || echo "0.1.0")

# Colors
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo -e "${BLUE}Tavor Go SDK - Available Commands${NC}"
	@echo -e "${BLUE}==================================${NC}"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

install: ## Install dependencies
	@echo -e "${BLUE}Installing dependencies...${NC}"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo -e "${GREEN}✓ Dependencies installed${NC}"

build: ## Build the library
	@echo -e "${BLUE}Building library...${NC}"
	$(GOBUILD) -v ./...
	@echo -e "${GREEN}✓ Build complete${NC}"

test: ## Run tests
	@echo -e "${BLUE}Running tests...${NC}"
	$(GOTEST) -v -race ./...

test-coverage: ## Run tests with coverage report
	@echo -e "${BLUE}Running tests with coverage...${NC}"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo -e "${GREEN}✓ Coverage report generated in coverage.html${NC}"
	@echo -e "${BLUE}Coverage summary:${NC}"
	@$(GOCMD) tool cover -func=coverage.out | tail -1

lint: ## Run linting checks
	@echo -e "${BLUE}Running go vet...${NC}"
	$(GOVET) ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo -e "${BLUE}Running golangci-lint...${NC}"; \
		$(GOLINT) run ./...; \
	else \
		echo -e "${YELLOW}golangci-lint not installed. Install with:${NC}"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi
	@echo -e "${GREEN}✓ Linting complete${NC}"

format: ## Format code
	@echo -e "${BLUE}Formatting code...${NC}"
	$(GOFMT) -w .
	$(GOMOD) tidy
	@echo -e "${GREEN}✓ Code formatted${NC}"

clean: ## Clean build artifacts and cache
	@echo -e "${BLUE}Cleaning...${NC}"
	$(GOCMD) clean -cache
	rm -f coverage.out coverage.html
	rm -rf vendor/
	@echo -e "${GREEN}✓ Clean complete${NC}"

examples: ## Build and test examples
	@echo -e "${BLUE}Building examples...${NC}"
	@for dir in examples/*/; do \
		if [ -f "$$dir/main.go" ]; then \
			echo -e "${BLUE}Building $$dir...${NC}"; \
			cd $$dir && $(GOBUILD) -o example . && cd ../..; \
		fi \
	done
	@echo -e "${GREEN}✓ Examples built${NC}"

check-version: ## Check current version
	@if [ ! -f $(VERSION_FILE) ]; then \
		echo -e "${YELLOW}Creating version file...${NC}"; \
		echo 'package tavor\n\n// Version is the current version of the SDK\nconst Version = "0.1.0"' > $(VERSION_FILE); \
	fi
	@echo -e "${BLUE}Current version: ${GREEN}$(CURRENT_VERSION)${NC}"

bump-patch: ## Bump patch version (0.0.X)
	@$(MAKE) _bump TYPE=patch

bump-minor: ## Bump minor version (0.X.0)
	@$(MAKE) _bump TYPE=minor

bump-major: ## Bump major version (X.0.0)
	@$(MAKE) _bump TYPE=major

_bump:
	@echo -e "${BLUE}Bumping $(TYPE) version from $(CURRENT_VERSION)...${NC}"
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{if("$(TYPE)"=="major") print $$1+1".0.0"; else if("$(TYPE)"=="minor") print $$1"."$$2+1".0"; else print $$1"."$$2"."$$3+1}'); \
	echo -e "${BLUE}New version: ${GREEN}$$NEW_VERSION${NC}"; \
	if [ -f $(VERSION_FILE) ]; then \
		sed -i.bak "s/$(CURRENT_VERSION)/$$NEW_VERSION/" $(VERSION_FILE) && rm $(VERSION_FILE).bak; \
	else \
		echo -e 'package tavor\n\n// Version is the current version of the SDK\nconst Version = "'$$NEW_VERSION'"' > $(VERSION_FILE); \
	fi; \
	echo -e "${GREEN}✓ Version bumped to $$NEW_VERSION${NC}"

tag: ## Create and push a git tag for the current version
	@echo -e "${BLUE}Creating tag v$(CURRENT_VERSION)...${NC}"
	git tag -a v$(CURRENT_VERSION) -m "Release v$(CURRENT_VERSION)"
	@echo -e "${GREEN}✓ Tag created${NC}"
	@echo -e "${BLUE}Pushing tag to origin...${NC}"
	git push origin v$(CURRENT_VERSION)
	@echo -e "${GREEN}✓ Tag pushed successfully${NC}"

index-module: ## Index the module in Go proxy
	@echo -e "${BLUE}Indexing module $(PACKAGE_NAME)@v$(CURRENT_VERSION) in Go proxy...${NC}"
	GOPROXY=proxy.golang.org go list -m $(PACKAGE_NAME)@v$(CURRENT_VERSION)
	@echo -e "${GREEN}✓ Module indexed successfully${NC}"

dev: install ## Set up development environment
	@echo -e "${GREEN}✓ Development environment ready!${NC}"
	@echo -e "${BLUE}Run tests with:${NC} make test"
	@echo -e "${BLUE}Format code with:${NC} make format"
	@echo -e "${BLUE}Run linting with:${NC} make lint"
	@echo -e "${BLUE}Build examples with:${NC} make examples"

pre-commit: format lint test ## Run all checks before committing
	@echo -e "${GREEN}✓ All pre-commit checks passed!${NC}"

release: pre-commit check-version ## Full release process
	@echo -e "${BLUE}Starting release process for version $(CURRENT_VERSION)...${NC}"
	@echo -e "${YELLOW}This will:${NC}"
	@echo "  1. Run all tests"
	@echo "  2. Create and push a git tag"
	@echo "  3. Index the module in Go proxy"
	@read -p "Continue? (y/N) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(MAKE) tag; \
		@echo -e "${BLUE}Waiting for tag to be available...${NC}"; \
		sleep 5; \
		$(MAKE) index-module; \
		echo -e "${GREEN}✓ Release complete!${NC}"; \
		echo -e "${YELLOW}Next steps:${NC}"; \
		echo "  1. Create a release on GitHub for v$(CURRENT_VERSION)"; \
		echo "  2. The module is now available at $(PACKAGE_NAME)@v$(CURRENT_VERSION)"; \
	fi

install-tools: ## Install development tools
	@echo -e "${BLUE}Installing development tools...${NC}"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo -e "${GREEN}✓ Tools installed${NC}"

bench: ## Run benchmarks
	@echo -e "${BLUE}Running benchmarks...${NC}"
	$(GOTEST) -bench=. -benchmem ./...

docs: ## Generate documentation
	@echo -e "${BLUE}Generating documentation...${NC}"
	@if command -v godoc >/dev/null 2>&1; then \
		echo -e "${GREEN}Starting godoc server at http://localhost:6060${NC}"; \
		echo -e "${GREEN}Package docs at: http://localhost:6060/pkg/$(PACKAGE_NAME)/${NC}"; \
		godoc -http=:6060; \
	else \
		echo -e "${YELLOW}godoc not installed. Install with:${NC}"; \
		echo "  go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

check-module: ## Check module dependencies
	@echo -e "${BLUE}Checking module dependencies...${NC}"
	$(GOMOD) verify
	$(GOGET) -u -t ./...
	$(GOMOD) tidy
	@echo -e "${GREEN}✓ Module dependencies updated${NC}"