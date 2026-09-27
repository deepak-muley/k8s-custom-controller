# Setup Guide

## Initial Setup

### 1. Install Dependencies

#### Go Dependencies
```bash
go mod download
go mod tidy
```

#### Pre-commit (for local development)
```bash
# Install pre-commit (one-time setup)
make pre-commit-install

# Or install manually:
pip install pre-commit
# or
brew install pre-commit

# Then install hooks:
make pre-commit-install
```

### 2. Build the Project
```bash
make build
```

### 3. Verify Installation
```bash
# Check binary
./bin/k8s-custom-controller --help

# Run tests
make test

# Check code quality
make fmt
make vet
make lint
```

## Pre-commit Setup

### Install Pre-commit Hooks

After cloning the repository, install pre-commit hooks:

```bash
make pre-commit-install
```

This will:
1. Install the pre-commit tool (if not already installed)
2. Install all git hooks defined in `.pre-commit-config.yaml`
3. Hooks will automatically run on `git commit`

### Pre-commit Checks

The pre-commit hooks include:
- **Code Formatting**: Automatically formats Go code (`make fmt`)
- **Go Vet**: Runs `go vet` on Go files
- **Unit Tests**: Runs unit tests (`make test-unit`)
- **Linting**: Runs golangci-lint
- **YAML/JSON Validation**: Validates YAML and JSON files
- **Helm Chart Linting**: Lints Helm charts when changed
- **File Checks**: Trailing whitespace, end-of-file, merge conflicts, etc.

### Running Pre-commit Manually

```bash
# Run pre-commit on all files
make pre-commit-run

# Run pre-commit on staged files only
pre-commit run

# Run specific hook
pre-commit run go-fmt
pre-commit run golangci-lint
```

### Updating Pre-commit Hooks

```bash
# Update to latest versions
make pre-commit-update

# Or manually
pre-commit autoupdate
```

### Skipping Pre-commit Hooks (Not Recommended)

If you need to skip pre-commit hooks for a specific commit:

```bash
git commit --no-verify -m "message"
```

**Warning**: Skipping hooks should only be done in exceptional circumstances.

## Development Workflow

### Before Committing

1. **Run pre-commit checks manually** (optional, they run automatically):
   ```bash
   make pre-commit-run
   ```

2. **Or run individual checks**:
   ```bash
   make fmt
   make vet
   make test
   make lint
   ```

### Before Pushing

Run pre-push checks (lighter than full CI):

```bash
make pre-push
```

This runs:
- `make fmt`
- `make vet`
- `make lint`
- `make test`
- `make build`
- `make helm-lint`
- `make pluto-detect-validate`

### Full CI Pipeline

To run the full CI pipeline locally (same as GitHub Actions):

```bash
make ci
```

This includes everything in `pre-push` plus:
- `make security-scan` (kubesec validation)
- `make pluto-detect-validate` (API version validation)

## Troubleshooting

### Pre-commit Not Running

1. Check if hooks are installed:
   ```bash
   ls -la .git/hooks/pre-commit
   ```

2. Reinstall hooks:
   ```bash
   make pre-commit-uninstall
   make pre-commit-install
   ```

### Pre-commit Hooks Failing

1. Run hooks manually to see detailed output:
   ```bash
   make pre-commit-run
   ```

2. Fix issues shown in the output

3. Re-run:
   ```bash
   git add .
   pre-commit run
   ```

### Pre-commit Taking Too Long

Some hooks can be disabled temporarily:

1. Edit `.pre-commit-config.yaml`
2. Comment out slow hooks
3. Or use `SKIP` environment variable:
   ```bash
   SKIP=golangci-lint git commit -m "message"
   ```

## Additional Tools

### Required for Full Development

- **Go 1.24+**: For building and testing (matches go.mod requirement)
- **Helm 3.12+**: For Helm chart operations
- **kubesec**: Installed via `make kubesec-install`
- **Pluto**: Installed via `make pluto-install`
- **pre-commit**: Installed via `make pre-commit-install`

### Optional Tools

- **golangci-lint**: For enhanced linting (installed automatically)
- **jq**: For better JSON parsing in Makefile scripts
- **Docker**: For building container images

## Quick Start

```bash
# 1. Clone the repository
git clone <repo-url>
cd k8s-custom-controller

# 2. Install Go dependencies
go mod download

# 3. Install pre-commit hooks (optional but recommended)
make pre-commit-install

# 4. Build the project
make build

# 5. Run tests
make test

# 6. You're ready to develop!
```

## Common Commands

```bash
# Build
make build

# Test
make test

# Format and check
make fmt
make vet
make lint

# Full CI checks
make ci

# Pre-push checks
make pre-push
```

