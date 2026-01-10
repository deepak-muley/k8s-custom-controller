# GitHub Actions CI/CD Pipeline

This directory contains GitHub Actions workflows for automated CI/CD.

## Workflows

### CI Pipeline (`ci.yml`)

Automated CI/CD pipeline that runs on every push and pull request.

#### Features

- **Automated Testing**: Runs tests, code formatting, vet, and linting
- **Multi-platform Build**: Builds binaries for Linux (amd64, arm64) and macOS
- **Docker Image Build**: Builds and pushes Docker images to GitHub Container Registry (GHCR)
- **Helm Chart**: Builds, lints, and validates Helm charts with kubesec
- **OCI Artifact Push**: Pushes Helm charts as OCI artifacts to GHCR
- **Security Scanning**: Runs kubesec scan with minimum score validation (default: 90)

#### Workflow Jobs

1. **test** - Runs tests, formatting checks, vet, and linter
2. **build** - Builds binaries for all platforms
3. **docker-build** - Builds and pushes Docker images to GHCR
4. **helm-chart** - Builds, lints, validates with kubesec, and packages Helm chart
5. **push-helm-chart** - Pushes Helm chart as OCI artifact to GHCR
6. **ci-summary** - Generates summary of all pipeline results

#### Triggers

- Push to `main`, `master`, or `develop` branches
- Pull requests to `main`, `master`, or `develop` branches
- Tags matching `v*` pattern (e.g., `v1.0.0`)

#### Artifacts

- Binary artifacts (Linux and macOS)
- Helm chart packages (.tgz)
- Kubesec scan reports (JSON)
- Docker images pushed to `ghcr.io/deepak-muley/k8s-custom-controller`

#### Using the Helm Chart from GHCR

After the pipeline runs, you can use the Helm chart:

```bash
# Authenticate to GHCR
helm registry login ghcr.io -u USERNAME -p TOKEN

# Install the chart (note: chart has -chart postfix with version)
helm install my-controller oci://ghcr.io/deepak-muley/k8s-custom-controller-chart:1.0.0 \
  --namespace default \
  --set controller.namespace=default \
  --set controller.gvks="apps/v1/Deployment,apps/v1/ReplicaSet"

# Or pull the chart first
helm pull oci://ghcr.io/deepak-muley/k8s-custom-controller-chart:1.0.0
```

#### Environment Variables

The workflow uses the following environment variables:

- `REGISTRY`: `ghcr.io`
- `IMAGE_NAME`: `${{ github.repository_owner }}/k8s-custom-controller`
- `HELM_CHART_NAME`: `k8s-custom-controller`
- `HELM_CHART_DIR`: `helm/k8s-custom-controller`
- `KUBESEC_MIN_SCORE`: `90`

#### Permissions

The workflow requires the following GitHub permissions:

- `contents: read` - To checkout code
- `packages: write` - To push Docker images and Helm charts to GHCR

These are automatically granted via `GITHUB_TOKEN` in GitHub Actions.

#### Kubesec Scanning

All Helm templates are scanned with kubesec to ensure security best practices:

- Minimum score requirement: 90/100
- Reports are generated in JSON format
- Failed scans will fail the CI pipeline

#### Versioning

- **Tagged releases**: Uses tag version (e.g., `v1.0.0` → version `1.0.0`)
- **Branch commits**: Uses commit SHA (e.g., `0.1.0-abc12345`)
- Chart version and appVersion are automatically updated

