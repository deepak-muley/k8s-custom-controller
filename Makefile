# Makefile for k8s-custom-controller

# Variables
BINARY_NAME ?= k8s-custom-controller
IMAGE_NAME ?= k8s-custom-controller
IMAGE_TAG ?= latest
IMAGE_REPO ?= docker.io/$(IMAGE_NAME)
HELM_CHART_NAME ?= k8s-custom-controller
HELM_CHART_DIR ?= helm/$(HELM_CHART_NAME)
KUBESEC_MIN_SCORE ?= 90
NAMESPACE ?= default

# Go variables
GO ?= go
GOFMT ?= gofmt
GOLANGCI_LINT ?= golangci-lint
GO_VERSION := $(shell $(GO) version | awk '{print $$3}')

# Build variables
BUILD_DIR ?= bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS ?= -w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

# Helm variables
HELM ?= helm
HELM_VERSION ?= 3.12.0
HELM_REGISTRY ?= ghcr.io
HELM_REPO ?= $(IMAGE_NAME)
HELM_CHART_VERSION ?= $(VERSION)

# Kubesec variables
KUBESEC ?= kubesec
KUBESEC_VERSION ?= 2.11.0

# Pluto variables
PLUTO ?= pluto
PLUTO_VERSION ?= 5.22.7
PLUTO_TARGET_K8S_VERSION ?= 1.29

.PHONY: all
all: clean fmt vet test build

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-linux
build-linux: ## Build the binary for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./main.go
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

.PHONY: build-darwin
build-darwin: ## Build the binary for macOS
	@echo "Building $(BINARY_NAME) for macOS..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./main.go
	@GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./main.go
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-*"

.PHONY: build-all
build-all: build-linux build-darwin ## Build binaries for all platforms

# Test targets
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@$(GO) test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@$(GO) test -v -short ./...

# Code quality targets
.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@echo "Code formatted"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@$(GO) vet ./...
	@echo "Go vet passed"

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v $(GOLANGCI_LINT) > /dev/null; then \
		$(GOLANGCI_LINT) run ./...; \
	else \
		echo "golangci-lint not found, installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
		$(GOLANGCI_LINT) run ./...; \
	fi

.PHONY: tidy
tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@$(GO) mod tidy
	@echo "Go modules tidied"

# Clean targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf helm/$(HELM_CHART_NAME)/charts
	@rm -rf helm/$(HELM_CHART_NAME)/*.tgz
	@echo "Clean complete"

.PHONY: clean-all
clean-all: clean ## Clean all artifacts including helm chart
	@rm -rf helm/

# Docker targets
.PHONY: docker-build
docker-build: build-linux ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .
	@docker tag $(IMAGE_REPO):$(IMAGE_TAG) $(IMAGE_REPO):latest
	@echo "Docker image built: $(IMAGE_REPO):$(IMAGE_TAG)"

.PHONY: docker-buildx
docker-buildx: ## Build Docker image with buildx (supports multi-platform)
	@echo "Building Docker image with buildx..."
	@if ! docker buildx ls | grep -q multiarch; then \
		docker buildx create --name multiarch --use || true; \
	fi
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t $(IMAGE_REPO):$(IMAGE_TAG) \
		-t $(IMAGE_REPO):latest \
		--load \
		.
	@echo "Docker image built with buildx: $(IMAGE_REPO):$(IMAGE_TAG)"

.PHONY: docker-buildx-push
docker-buildx-push: ## Build and push Docker image with buildx (multi-platform) - requires IMAGE_REPO, IMAGE_TAG
	@echo "Building and pushing Docker image with buildx (multi-platform)..."
	@if [ -z "$(IMAGE_REPO)" ] || [ -z "$(IMAGE_TAG)" ]; then \
		echo "ERROR: IMAGE_REPO and IMAGE_TAG must be set"; \
		echo "Usage: make docker-buildx-push IMAGE_REPO=ghcr.io/owner/image IMAGE_TAG=1.0.0"; \
		exit 1; \
	fi
	@echo "Creating buildx builder if needed..."
	@docker buildx create --name multiarch --driver docker-container --use 2>/dev/null || \
		docker buildx use multiarch 2>/dev/null || \
		docker buildx inspect --bootstrap || true
	@echo "Building and pushing multi-platform image: $(IMAGE_REPO):$(IMAGE_TAG)"
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t $(IMAGE_REPO):$(IMAGE_TAG) \
		-t $(IMAGE_REPO):latest \
		--push \
		--cache-from type=gha \
		--cache-to type=gha,mode=max \
		.
	@echo "Docker image built and pushed successfully: $(IMAGE_REPO):$(IMAGE_TAG)"

.PHONY: docker-push
docker-push: docker-build ## Push Docker image
	@echo "Pushing Docker image..."
	@docker push $(IMAGE_REPO):$(IMAGE_TAG)
	@docker push $(IMAGE_REPO):latest
	@echo "Docker image pushed"

# Helm chart targets
.PHONY: helm-create
helm-create: ## Create Helm chart structure
	@echo "Creating Helm chart..."
	@mkdir -p $(HELM_CHART_DIR)
	@mkdir -p $(HELM_CHART_DIR)/templates
	@mkdir -p $(HELM_CHART_DIR)/charts
	@echo "Helm chart structure created at $(HELM_CHART_DIR)"

.PHONY: helm-lint
helm-lint: helm-chart ## Lint Helm chart
	@echo "Linting Helm chart..."
	@$(HELM) lint $(HELM_CHART_DIR)
	@echo "Helm chart linted successfully"

.PHONY: helm-template
helm-template: helm-chart ## Render Helm templates to stdout
	@echo "Rendering Helm templates..."
	@$(HELM) template $(HELM_CHART_NAME) $(HELM_CHART_DIR) \
		--namespace $(NAMESPACE) \
		--set image.repository=$(IMAGE_REPO) \
		--set image.tag=$(IMAGE_TAG)

.PHONY: helm-template-dir
helm-template-dir: helm-chart ## Render Helm templates to directory
	@echo "Rendering Helm templates to rendered/ directory..."
	@mkdir -p rendered
	@$(HELM) template $(HELM_CHART_NAME) $(HELM_CHART_DIR) \
		--namespace $(NAMESPACE) \
		--set image.repository=$(IMAGE_REPO) \
		--set image.tag=$(IMAGE_TAG) \
		--output-dir rendered/
	@echo "Templates rendered to rendered/ directory"

.PHONY: helm-package
helm-package: helm-chart ## Package Helm chart
	@echo "Packaging Helm chart..."
	@$(HELM) package $(HELM_CHART_DIR) --destination $(HELM_CHART_DIR)/charts
	@echo "Helm chart packaged"

.PHONY: helm-push-oci
helm-push-oci: helm-package ## Push Helm chart as OCI artifact with -chart postfix (requires HELM_REGISTRY, HELM_REPO, HELM_CHART_VERSION)
	@echo "Pushing Helm chart as OCI artifact..."
	@if [ -z "$(HELM_REGISTRY)" ] || [ -z "$(HELM_REPO)" ]; then \
		echo "ERROR: HELM_REGISTRY and HELM_REPO must be set"; \
		echo "Usage: make helm-push-oci HELM_REGISTRY=ghcr.io HELM_REPO=owner/chart-name HELM_CHART_VERSION=1.0.0"; \
		exit 1; \
	fi
	@chart_file=$$(ls -t $(HELM_CHART_DIR)/charts/$(HELM_CHART_NAME)-*.tgz 2>/dev/null | head -1); \
	if [ -z "$$chart_file" ]; then \
		echo "ERROR: Chart package not found. Run 'make helm-package' first"; \
		exit 1; \
	fi; \
	if [ -n "$(HELM_CHART_VERSION)" ]; then \
		chart_version="$(HELM_CHART_VERSION)"; \
	else \
		chart_version=$$(basename $$chart_file .tgz | sed 's/.*-\([0-9].*\)/\1/'); \
	fi; \
	# Add -chart postfix with version to the OCI artifact name: owner/repo-chart:version \
	oci_repo="$(HELM_REGISTRY)/$(HELM_REPO)-chart"; \
	oci_ref="$$oci_repo:$$chart_version"; \
	echo "Pushing chart $$chart_file to oci://$$oci_ref"; \
	# Use helm chart save/push for explicit version tagging (Helm 3.8+) \
	# This allows us to specify the exact version tag with -chart postfix \
	if $(HELM) chart save $$chart_file $$oci_ref 2>/dev/null && $(HELM) chart push $$oci_ref 2>/dev/null; then \
		echo "Chart pushed successfully using 'helm chart save/push'"; \
		echo "Helm chart pushed as OCI artifact: oci://$$oci_ref"; \
	elif $(HELM) push $$chart_file oci://$$oci_repo 2>/dev/null; then \
		echo "Chart pushed using 'helm push', but version may not match"; \
		echo "Warning: For explicit version tagging, ensure Helm 3.8+ is installed"; \
		echo "Helm chart location: oci://$$oci_repo"; \
	else \
		echo "ERROR: Failed to push chart to OCI registry"; \
		echo "Attempted to push to: oci://$$oci_ref"; \
		echo "Ensure you are logged in: helm registry login $(HELM_REGISTRY)"; \
		exit 1; \
	fi

.PHONY: helm-dependency
helm-dependency: helm-chart ## Update Helm chart dependencies
	@echo "Updating Helm chart dependencies..."
	@$(HELM) dependency update $(HELM_CHART_DIR)
	@echo "Helm chart dependencies updated"

.PHONY: helm-chart
helm-chart: helm-create ## Ensure Helm chart exists (creates if missing)
	@if [ ! -f "$(HELM_CHART_DIR)/Chart.yaml" ]; then \
		echo "Helm chart not found, creating..."; \
		$(MAKE) helm-create; \
	fi

# Kubesec targets
.PHONY: kubesec-install
kubesec-install: ## Install kubesec
	@echo "Installing kubesec..."
	@if command -v $(KUBESEC) > /dev/null; then \
		echo "kubesec already installed"; \
	else \
		echo "Downloading kubesec..."; \
		curl -sSLo /tmp/kubesec.tar.gz "https://github.com/controlplaneio/kubesec/releases/download/v$(KUBESEC_VERSION)/kubesec_$(KUBESEC_VERSION)_linux_amd64.tar.gz"; \
		mkdir -p /tmp/kubesec; \
		tar -xzf /tmp/kubesec.tar.gz -C /tmp/kubesec; \
		sudo mv /tmp/kubesec/kubesec /usr/local/bin/kubesec; \
		chmod +x /usr/local/bin/kubesec; \
		rm -rf /tmp/kubesec /tmp/kubesec.tar.gz; \
		echo "kubesec installed"; \
	fi

.PHONY: kubesec-scan
kubesec-scan: helm-template-dir kubesec-install ## Scan rendered Helm templates with kubesec
	@echo "Scanning rendered templates with kubesec (minimum score: $(KUBESEC_MIN_SCORE))..."
	@min_score=$(KUBESEC_MIN_SCORE); \
	failed=0; \
	file_count=0; \
	for file in rendered/$(HELM_CHART_NAME)/templates/*.yaml rendered/$(HELM_CHART_NAME)/templates/*.yml; do \
		if [ -f "$$file" ]; then \
			file_count=$$((file_count + 1)); \
			filename=$$(basename $$file); \
			echo "Scanning $$filename..."; \
			result=$$($(KUBESEC) scan $$file 2>/dev/null || echo '[]'); \
			if [ -z "$$result" ] || [ "$$result" = "[]" ]; then \
				echo "  WARNING: No resources found or failed to scan"; \
				continue; \
			fi; \
			# Handle both single object and array responses \
			if command -v jq > /dev/null 2>&1; then \
				scores=$$(echo $$result | jq -r 'if type=="array" then .[].score else .score end' 2>/dev/null || echo "0"); \
				for score in $$scores; do \
					if [ "$$score" != "null" ] && [ -n "$$score" ]; then \
						printf "    Score: %3s (min: %3s)" "$$score" "$$min_score"; \
						if [ "$$score" -ge "$$min_score" ] 2>/dev/null; then \
							echo " ✓ PASS"; \
						else \
							echo " ✗ FAIL"; \
							failed=1; \
						fi; \
					fi; \
				done; \
			else \
				score=$$(echo $$result | grep -oE '"score"[[:space:]]*:[[:space:]]*[0-9]+' | grep -oE '[0-9]+' | head -1 || echo "0"); \
				if [ -z "$$score" ] || [ "$$score" = "null" ]; then \
					score=0; \
				fi; \
				printf "    Score: %3s (min: %3s)" "$$score" "$$min_score"; \
				if [ "$$score" -ge "$$min_score" ] 2>/dev/null; then \
					echo " ✓ PASS"; \
				else \
					echo " ✗ FAIL"; \
					failed=1; \
				fi; \
			fi; \
		fi; \
	done; \
	if [ "$$file_count" -eq 0 ]; then \
		echo "ERROR: No template files found to scan"; \
		exit 1; \
	fi; \
	if [ "$$failed" -eq 1 ]; then \
		echo ""; \
		echo "Kubesec scan FAILED: One or more resources scored below minimum $(KUBESEC_MIN_SCORE)"; \
		echo "Run 'make kubesec-scan-json' for detailed reports"; \
		exit 1; \
	else \
		echo ""; \
		echo "Kubesec scan PASSED: All resources meet minimum score of $(KUBESEC_MIN_SCORE)"; \
	fi

.PHONY: kubesec-scan-json
kubesec-scan-json: helm-template-dir kubesec-install ## Scan rendered Helm templates with kubesec (JSON output)
	@echo "Scanning rendered templates with kubesec (JSON output)..."
	@mkdir -p kubesec-reports
	@for file in rendered/$(HELM_CHART_NAME)/templates/*.yaml rendered/$(HELM_CHART_NAME)/templates/*.yml; do \
		if [ -f "$$file" ]; then \
			filename=$$(basename $$file .yaml); \
			filename=$$(basename $$filename .yml); \
			echo "Scanning $$file..."; \
			$(KUBESEC) scan $$file > kubesec-reports/$$filename.json 2>/dev/null || true; \
		fi; \
	done; \
	echo "Kubesec reports saved to kubesec-reports/ directory"

.PHONY: kubesec-validate
kubesec-validate: helm-template-dir kubesec-install ## Validate all rendered templates meet kubesec score requirement
	@echo "Validating kubesec scores (minimum: $(KUBESEC_MIN_SCORE))..."
	@min_score=$(KUBESEC_MIN_SCORE); \
	overall_failed=0; \
	file_count=0; \
	total_resources=0; \
	passed_resources=0; \
	for file in rendered/$(HELM_CHART_NAME)/templates/*.yaml rendered/$(HELM_CHART_NAME)/templates/*.yml; do \
		if [ -f "$$file" ]; then \
			file_count=$$((file_count + 1)); \
			filename=$$(basename $$file); \
			result=$$($(KUBESEC) scan $$file 2>/dev/null || echo '[]'); \
			if [ -z "$$result" ] || [ "$$result" = "[]" ]; then \
				printf "%-50s (no resources to scan)\n" "$$filename"; \
				continue; \
			fi; \
			# Handle both single object and array responses \
			if command -v jq > /dev/null 2>&1; then \
				resource_count=$$(echo $$result | jq -r 'if type=="array" then length else 1 end' 2>/dev/null || echo "1"); \
				scores=$$(echo $$result | jq -r 'if type=="array" then .[].score else .score end' 2>/dev/null || echo "0"); \
				resource_num=1; \
				for score in $$scores; do \
					total_resources=$$((total_resources + 1)); \
					if [ "$$score" != "null" ] && [ -n "$$score" ] && [ "$$score" != "" ]; then \
						if [ "$$resource_count" -gt 1 ]; then \
							printf "%-50s [%d] Score: %3s (min: %3s)" "$$filename" "$$resource_num" "$$score" "$$min_score"; \
						else \
							printf "%-50s Score: %3s (min: %3s)" "$$filename" "$$score" "$$min_score"; \
						fi; \
						if [ "$$score" -ge "$$min_score" ] 2>/dev/null; then \
							echo " ✓ PASS"; \
							passed_resources=$$((passed_resources + 1)); \
						else \
							echo " ✗ FAIL"; \
							overall_failed=1; \
						fi; \
						resource_num=$$((resource_num + 1)); \
					fi; \
				done; \
			else \
				total_resources=$$((total_resources + 1)); \
				score=$$(echo $$result | grep -oE '"score"[[:space:]]*:[[:space:]]*[0-9]+' | grep -oE '[0-9]+' | head -1 || echo "0"); \
				if [ -z "$$score" ] || [ "$$score" = "null" ] || [ "$$score" = "" ]; then \
					score=0; \
				fi; \
				printf "%-50s Score: %3s (min: %3s)" "$$filename" "$$score" "$$min_score"; \
				if [ "$$score" -ge "$$min_score" ] 2>/dev/null; then \
					echo " ✓ PASS"; \
					passed_resources=$$((passed_resources + 1)); \
				else \
					echo " ✗ FAIL"; \
					overall_failed=1; \
				fi; \
			fi; \
		fi; \
	done; \
	if [ "$$file_count" -eq 0 ]; then \
		echo "ERROR: No template files found to scan"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "Summary: $$passed_resources/$$total_resources resources passed (minimum score: $(KUBESEC_MIN_SCORE))"; \
	if [ "$$overall_failed" -eq 1 ]; then \
		echo "ERROR: One or more resources scored below minimum $(KUBESEC_MIN_SCORE)"; \
		echo "Run 'make kubesec-scan-json' to see detailed reports"; \
		exit 1; \
	else \
		echo "SUCCESS: All resources meet minimum kubesec score of $(KUBESEC_MIN_SCORE)"; \
	fi

# Pluto targets (API version deprecation detection)
.PHONY: pluto-install
pluto-install: ## Install pluto for detecting deprecated Kubernetes apiVersions
	@echo "Installing pluto..."
	@if command -v $(PLUTO) > /dev/null; then \
		echo "pluto already installed"; \
		$(PLUTO) version || echo "pluto found but version check failed"; \
	else \
		echo "Downloading pluto v$(PLUTO_VERSION)..."; \
		OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
		ARCH=$$(uname -m); \
		if [ "$$ARCH" = "x86_64" ]; then \
			ARCH="amd64"; \
		elif [ "$$ARCH" = "arm64" ] || [ "$$ARCH" = "aarch64" ]; then \
			ARCH="arm64"; \
		fi; \
		if [ "$$OS" = "darwin" ]; then \
			OS="darwin"; \
		elif [ "$$OS" = "linux" ]; then \
			OS="linux"; \
		else \
			echo "ERROR: Unsupported OS: $$OS"; \
			exit 1; \
		fi; \
		mkdir -p $(BUILD_DIR); \
		PLUTO_BIN="$(BUILD_DIR)/pluto"; \
		# Try versioned release first \
		PLUTO_URL="https://github.com/FairwindsOps/pluto/releases/download/v$(PLUTO_VERSION)/pluto_$(PLUTO_VERSION)_$${OS}_$${ARCH}.tar.gz"; \
		echo "Downloading from: $$PLUTO_URL"; \
		if curl -sSL -f -o /tmp/pluto.tar.gz "$$PLUTO_URL"; then \
			tar -xzf /tmp/pluto.tar.gz -C /tmp 2>/dev/null || true; \
			if [ -f "/tmp/pluto" ]; then \
				mv /tmp/pluto $$PLUTO_BIN; \
			elif [ -f "/tmp/pluto_$${OS}_$${ARCH}/pluto" ]; then \
				mv /tmp/pluto_$${OS}_$${ARCH}/pluto $$PLUTO_BIN; \
			else \
				echo "Trying direct binary download..."; \
				curl -sSL -f -o $$PLUTO_BIN "https://github.com/FairwindsOps/pluto/releases/download/v$(PLUTO_VERSION)/pluto-$${OS}-$${ARCH}" || \
				curl -sSL -f -o $$PLUTO_BIN "https://github.com/FairwindsOps/pluto/releases/latest/download/pluto-$${OS}-$${ARCH}" || { \
					echo "ERROR: Failed to download pluto"; \
					rm -rf /tmp/pluto* /tmp/pluto.tar.gz; \
					exit 1; \
				}; \
			fi; \
		else \
			echo "Versioned download failed, trying latest release..."; \
			curl -sSL -f -o /tmp/pluto.tar.gz "https://github.com/FairwindsOps/pluto/releases/latest/download/pluto_$${OS}_$${ARCH}.tar.gz" || \
			curl -sSL -f -o $$PLUTO_BIN "https://github.com/FairwindsOps/pluto/releases/latest/download/pluto-$${OS}-$${ARCH}" || { \
				echo "ERROR: Failed to download pluto from any source"; \
				exit 1; \
			}; \
			if [ -f "/tmp/pluto.tar.gz" ]; then \
				tar -xzf /tmp/pluto.tar.gz -C /tmp 2>/dev/null; \
				find /tmp -name "pluto" -type f -exec mv {} $$PLUTO_BIN \; 2>/dev/null || true; \
			fi; \
		fi; \
		chmod +x $$PLUTO_BIN; \
		sudo mv $$PLUTO_BIN /usr/local/bin/pluto 2>/dev/null || { \
			echo "Note: Installing to $(BUILD_DIR) (sudo not available)"; \
			mv $$PLUTO_BIN $(BUILD_DIR)/pluto; \
			export PATH="$$PATH:$(shell pwd)/$(BUILD_DIR)"; \
		}; \
		rm -rf /tmp/pluto* /tmp/pluto.tar.gz; \
		if command -v pluto > /dev/null 2>&1 || [ -f "$(BUILD_DIR)/pluto" ]; then \
			pluto version 2>/dev/null || $(BUILD_DIR)/pluto version 2>/dev/null || echo "pluto installed (version check skipped)"; \
			echo "pluto installed successfully"; \
		else \
			echo "ERROR: pluto binary not found after installation"; \
			exit 1; \
		fi; \
	fi

.PHONY: pluto-detect-files
pluto-detect-files: pluto-install ## Detect deprecated apiVersions in rendered Helm templates
	@echo "Detecting deprecated Kubernetes apiVersions in rendered templates..."
	@if [ ! -d "rendered/$(HELM_CHART_NAME)/templates" ]; then \
		echo "Rendered templates not found. Running helm-template-dir first..."; \
		$(MAKE) helm-template-dir; \
	fi
	@$(PLUTO) detect-files -d rendered/$(HELM_CHART_NAME)/templates \
		--target-version k8s=$(PLUTO_TARGET_K8S_VERSION) \
		--output wide \
		--ignore-deprecations=false \
		--ignore-removals=false || \
		$(PLUTO) detect-files -d rendered/$(HELM_CHART_NAME)/templates \
		--target-version k8s=$(PLUTO_TARGET_K8S_VERSION) || true

.PHONY: pluto-detect-charts
pluto-detect-charts: pluto-install helm-chart ## Detect deprecated apiVersions in Helm chart templates
	@echo "Detecting deprecated Kubernetes apiVersions in Helm chart..."
	@$(PLUTO) detect-helm -d $(HELM_CHART_DIR) \
		--target-version k8s=$(PLUTO_TARGET_K8S_VERSION) \
		--output wide \
		--ignore-deprecations=false \
		--ignore-removals=false || \
		$(PLUTO) detect-helm -d $(HELM_CHART_DIR) \
		--target-version k8s=$(PLUTO_TARGET_K8S_VERSION) || true

.PHONY: pluto-detect-validate
pluto-detect-validate: pluto-install helm-template-dir ## Validate no deprecated apiVersions found (fails on deprecations)
	@echo "Validating Kubernetes apiVersions (target K8s version: $(PLUTO_TARGET_K8S_VERSION))..."
	@if [ ! -d "rendered/$(HELM_CHART_NAME)/templates" ]; then \
		echo "Rendered templates not found. Running helm-template-dir first..."; \
		$(MAKE) helm-template-dir; \
	fi
	@echo "Checking rendered Helm templates..."
	@if $(PLUTO) detect-files -d rendered/$(HELM_CHART_NAME)/templates \
		--target-version k8s=$(PLUTO_TARGET_K8S_VERSION) \
		--output wide \
		--ignore-deprecations=false \
		--ignore-removals=false 2>&1 | grep -qE "(DEPRECATED|REMOVED)"; then \
		echo ""; \
		echo "ERROR: Deprecated or removed Kubernetes apiVersions found!"; \
		$(PLUTO) detect-files -d rendered/$(HELM_CHART_NAME)/templates \
			--target-version k8s=$(PLUTO_TARGET_K8S_VERSION) \
			--output wide \
			--ignore-deprecations=false \
			--ignore-removals=false; \
		exit 1; \
	else \
		echo "✓ No deprecated or removed apiVersions found"; \
	fi
	@echo "Checking Helm chart templates..."
	@if $(PLUTO) detect-helm -d $(HELM_CHART_DIR) \
		--target-version k8s=$(PLUTO_TARGET_K8S_VERSION) \
		--output wide \
		--ignore-deprecations=false \
		--ignore-removals=false 2>&1 | grep -qE "(DEPRECATED|REMOVED)"; then \
		echo ""; \
		echo "ERROR: Deprecated or removed Kubernetes apiVersions found in Helm chart!"; \
		$(PLUTO) detect-helm -d $(HELM_CHART_DIR) \
			--target-version k8s=$(PLUTO_TARGET_K8S_VERSION) \
			--output wide \
			--ignore-deprecations=false \
			--ignore-removals=false; \
		exit 1; \
	else \
		echo "✓ No deprecated or removed apiVersions found in Helm chart"; \
	fi
	@echo "Pluto validation passed successfully"

.PHONY: pluto-detect-all
pluto-detect-all: pluto-detect-charts pluto-detect-files ## Detect deprecated apiVersions in both charts and rendered templates
	@echo "Pluto API version detection completed"

# Combined targets
.PHONY: helm-build
helm-build: helm-chart helm-package ## Build and package Helm chart
	@echo "Helm chart built and packaged"

.PHONY: security-scan
security-scan: helm-template-dir kubesec-validate pluto-detect-validate ## Run full security scan (render templates, validate with kubesec, and check for deprecated apiVersions)
	@echo "Security scan completed successfully"

.PHONY: helm-build-push
helm-build-push: helm-package helm-push-oci ## Build, package and push Helm chart as OCI artifact
	@echo "Helm chart built, packaged and pushed"

.PHONY: pre-push
pre-push: fmt vet lint test build helm-lint pluto-detect-validate ## Run checks before pushing (lighter than full CI)
	@echo "Pre-push checks completed successfully"

.PHONY: ci
ci: fmt vet lint test build helm-lint security-scan pluto-detect-validate ## Run full CI pipeline (includes security scan and API version validation)
	@echo "CI pipeline completed successfully"

.PHONY: install
install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install -ldflags "$(LDFLAGS)" ./main.go
	@echo "Installed: $$(go env GOPATH)/bin/$(BINARY_NAME)"

# Devbox variables
DEVBOX ?= devbox
DEVBOX_INSTALLED := $(shell command -v $(DEVBOX) 2>/dev/null)

# Pre-commit hooks targets
PRE_COMMIT ?= pre-commit

.PHONY: pre-commit-install
pre-commit-install: ## Install pre-commit hooks
	@echo "Installing pre-commit..."
	@if command -v $(PRE_COMMIT) > /dev/null; then \
		echo "pre-commit already installed"; \
		$(PRE_COMMIT) --version; \
	else \
		echo "Installing pre-commit via pip..."; \
		if command -v pip3 > /dev/null; then \
			pip3 install --user pre-commit || pip3 install pre-commit; \
		elif command -v pip > /dev/null; then \
			pip install --user pre-commit || pip install pre-commit; \
		else \
			echo "ERROR: pip or pip3 not found. Please install pre-commit manually:"; \
			echo "  pip install pre-commit"; \
			echo "  or"; \
			echo "  brew install pre-commit"; \
			exit 1; \
		fi; \
		echo "pre-commit installed"; \
	fi
	@echo "Installing pre-commit hooks..."
	@$(PRE_COMMIT) install || \
		$(HOME)/.local/bin/pre-commit install || \
		$$(python3 -m site --user-base)/bin/pre-commit install || \
		echo "WARNING: Failed to install pre-commit hooks. Run manually: pre-commit install"
	@echo "Pre-commit hooks installed successfully"

.PHONY: pre-commit-uninstall
pre-commit-uninstall: ## Uninstall pre-commit hooks
	@echo "Uninstalling pre-commit hooks..."
	@if command -v $(PRE_COMMIT) > /dev/null; then \
		$(PRE_COMMIT) uninstall; \
	else \
		$(HOME)/.local/bin/pre-commit uninstall || \
		$$(python3 -m site --user-base)/bin/pre-commit uninstall || \
		echo "pre-commit not found"; \
	fi
	@echo "Pre-commit hooks uninstalled"

.PHONY: pre-commit-run
pre-commit-run: ## Run pre-commit hooks on all files
	@echo "Running pre-commit hooks on all files..."
	@if command -v $(PRE_COMMIT) > /dev/null; then \
		$(PRE_COMMIT) run --all-files; \
	else \
		$(HOME)/.local/bin/pre-commit run --all-files || \
		$$(python3 -m site --user-base)/bin/pre-commit run --all-files || \
		echo "ERROR: pre-commit not found. Run 'make pre-commit-install' first"; \
		exit 1; \
	fi

.PHONY: pre-commit-update
pre-commit-update: ## Update pre-commit hooks to latest versions
	@echo "Updating pre-commit hooks..."
	@if command -v $(PRE_COMMIT) > /dev/null; then \
		$(PRE_COMMIT) autoupdate; \
	else \
		$(HOME)/.local/bin/pre-commit autoupdate || \
		$$(python3 -m site --user-base)/bin/pre-commit autoupdate || \
		echo "ERROR: pre-commit not found. Run 'make pre-commit-install' first"; \
		exit 1; \
	fi
	@echo "Pre-commit hooks updated"

.PHONY: pre-commit
pre-commit: pre-commit-install pre-commit-run ## Install and run pre-commit hooks

# Devbox targets
.PHONY: devbox-install
devbox-install: ## Show devbox installation instructions
	@echo "Checking for devbox..."
	@if command -v $(DEVBOX) > /dev/null 2>&1; then \
		echo "✅ devbox already installed at: $$(command -v $(DEVBOX))"; \
		$(DEVBOX) --help 2>/dev/null | head -3 || echo "   (devbox is available)"; \
	elif [ -f "$(HOME)/.local/bin/devbox" ]; then \
		echo "✅ devbox found at ~/.local/bin/devbox"; \
	elif [ -f "$(HOME)/.devbox/bin/devbox" ]; then \
		echo "✅ devbox found at ~/.devbox/bin/devbox"; \
	else \
		echo "⚠️  devbox not found"; \
		echo ""; \
		echo "Install devbox:"; \
		echo "  macOS:"; \
		echo "    brew tap jetpack-io/devbox"; \
		echo "    brew install devbox"; \
		echo ""; \
		echo "  Linux:"; \
		echo "    curl -fsSL https://get.jetpack.io/devbox | bash"; \
		echo ""; \
		echo "  Or visit: https://www.jetpack.io/devbox/docs/installing_devbox/"; \
		echo ""; \
		echo "After installation, verify with: devbox --help"; \
		exit 1; \
	fi

.PHONY: devbox-shell
devbox-shell: ## Start devbox shell with all development tools
	@echo "Starting devbox shell..."
	@if command -v $(DEVBOX) > /dev/null 2>&1; then \
		echo "Entering devbox shell..."; \
		echo "Type 'exit' or press Ctrl+D to leave the shell"; \
		$(DEVBOX) shell; \
	elif [ -f "$(HOME)/.local/bin/devbox" ]; then \
		echo "Using devbox from ~/.local/bin"; \
		$(HOME)/.local/bin/devbox shell; \
	elif [ -f "$(HOME)/.devbox/bin/devbox" ]; then \
		echo "Using devbox from ~/.devbox/bin"; \
		$(HOME)/.devbox/bin/devbox shell; \
	else \
		echo "ERROR: devbox not found."; \
		echo ""; \
		echo "Install devbox:"; \
		echo "  macOS:    brew install jetpack-io/devbox/devbox"; \
		echo "  Linux:    curl -fsSL https://get.jetpack.io/devbox | bash"; \
		echo "  Or visit: https://www.jetpack.io/devbox/docs/installing_devbox/"; \
		echo ""; \
		echo "After installation, verify with: devbox --version"; \
		exit 1; \
	fi

.PHONY: devbox-init
devbox-init: ## Initialize devbox (creates devbox.json if missing)
	@echo "Initializing devbox..."
	@if command -v $(DEVBOX) > /dev/null 2>&1 || [ -f "$(HOME)/.local/bin/devbox" ] || [ -f "$(HOME)/.devbox/bin/devbox" ]; then \
		DEVBOX_CMD=$$(command -v $(DEVBOX) 2>/dev/null || echo "$(HOME)/.local/bin/devbox" || echo "$(HOME)/.devbox/bin/devbox"); \
		if [ ! -f "devbox.json" ]; then \
			$$DEVBOX_CMD init || echo "devbox.json already exists or initialization failed"; \
		else \
			echo "devbox.json already exists"; \
		fi; \
	else \
		echo "ERROR: devbox not found. Install devbox first:"; \
		echo "  macOS: brew install jetpack-io/devbox/devbox"; \
		echo "  Linux: curl -fsSL https://get.jetpack.io/devbox | bash"; \
		exit 1; \
	fi

.PHONY: devbox-run
devbox-run: ## Run a command in devbox environment (usage: make devbox-run CMD="make build")
	@if [ -z "$(CMD)" ]; then \
		echo "ERROR: CMD not specified"; \
		echo "Usage: make devbox-run CMD=\"make build\""; \
		echo "Example: make devbox-run CMD=\"make test\""; \
		exit 1; \
	fi
	@echo "Running command in devbox: $(CMD)"
	@if command -v $(DEVBOX) > /dev/null 2>&1; then \
		$(DEVBOX) run $(CMD); \
	elif [ -f "$(HOME)/.local/bin/devbox" ]; then \
		$(HOME)/.local/bin/devbox run $(CMD); \
	elif [ -f "$(HOME)/.devbox/bin/devbox" ]; then \
		$(HOME)/.devbox/bin/devbox run $(CMD); \
	else \
		echo "ERROR: devbox not found. Run 'make devbox-install' for installation instructions"; \
		exit 1; \
	fi

.PHONY: devbox-update
devbox-update: ## Update devbox packages to latest versions
	@echo "Updating devbox packages..."
	@if command -v $(DEVBOX) > /dev/null 2>&1; then \
		$(DEVBOX) add --latest go helm kubectl jq git curl bash python3; \
		echo "✅ Devbox packages updated"; \
	elif [ -f "$(HOME)/.local/bin/devbox" ]; then \
		$(HOME)/.local/bin/devbox add --latest go helm kubectl jq git curl bash python3; \
		echo "✅ Devbox packages updated"; \
	elif [ -f "$(HOME)/.devbox/bin/devbox" ]; then \
		$(HOME)/.devbox/bin/devbox add --latest go helm kubectl jq git curl bash python3; \
		echo "✅ Devbox packages updated"; \
	else \
		echo "ERROR: devbox not found. Run 'make devbox-install' for installation instructions"; \
		exit 1; \
	fi

.PHONY: devbox-info
devbox-info: ## Show devbox info and installed packages
	@if command -v $(DEVBOX) > /dev/null 2>&1; then \
		echo "Devbox location: $$(command -v $(DEVBOX))"; \
		echo ""; \
		echo "Configured packages (from devbox.json):"; \
		if [ -f "devbox.json" ]; then \
			jq -r '.packages[]' devbox.json 2>/dev/null | sed 's/^/  - /' || \
			grep -o '"[^"]*@[^"]*"' devbox.json 2>/dev/null | sed 's/"//g' | sed 's/^/  - /' || \
			echo "  (Unable to parse devbox.json - install jq for better parsing)"; \
		else \
			echo "  devbox.json not found"; \
		fi; \
		echo ""; \
		echo "Run 'make devbox-shell' to enter the devbox environment"; \
	elif [ -f "$(HOME)/.local/bin/devbox" ]; then \
		echo "Devbox location: ~/.local/bin/devbox"; \
		if [ -f "devbox.json" ]; then \
			jq -r '.packages[]' devbox.json 2>/dev/null | sed 's/^/  - /' || \
			grep -o '"[^"]*@[^"]*"' devbox.json 2>/dev/null | sed 's/"//g' | sed 's/^/  - /' || \
			echo "  (Unable to parse devbox.json)"; \
		fi; \
	elif [ -f "$(HOME)/.devbox/bin/devbox" ]; then \
		echo "Devbox location: ~/.devbox/bin/devbox"; \
		if [ -f "devbox.json" ]; then \
			jq -r '.packages[]' devbox.json 2>/dev/null | sed 's/^/  - /' || \
			grep -o '"[^"]*@[^"]*"' devbox.json 2>/dev/null | sed 's/"//g' | sed 's/^/  - /' || \
			echo "  (Unable to parse devbox.json)"; \
		fi; \
	else \
		echo "⚠️  devbox not installed. Run 'make devbox-install' for installation instructions"; \
	fi

# Default target
.DEFAULT_GOAL := help

