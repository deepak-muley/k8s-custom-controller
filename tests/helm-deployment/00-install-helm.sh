#!/bin/bash
# KUTTL Pre-Install Script: Install Helm Chart
# This script installs the Helm chart before KUTTL tests run
# KUTTL will execute this script automatically if named correctly
#
# This demonstrates installing the ENTIRE Helm chart (not individual YAMLs)
# The script performs a full "helm install" with all resources

set -euo pipefail

CHART_NAME="${CHART_NAME:-k8s-custom-controller}"
RELEASE_NAME="${RELEASE_NAME:-k8s-custom-controller}"
NAMESPACE="${NAMESPACE:-kuttl-test-helm}"
# Try packaged chart first, then chart directory
CHART_PATH="${CHART_PATH:-}"
IMAGE_REPO="${IMAGE_REPO:-k8s-custom-controller}"
IMAGE_TAG="${IMAGE_TAG:-test}"
CONTROLLER_NAMESPACE="${CONTROLLER_NAMESPACE:-kuttl-test-controller}"
GVKS="${GVKS:-apps/v1/Deployment}"

# Find chart path (prefer packaged .tgz, fallback to directory)
if [ -z "${CHART_PATH}" ]; then
    # Try to find packaged chart in charts directory
    CHART_TGZ=$(find ../../../helm -name "*.tgz" -type f 2>/dev/null | head -1)
    if [ -n "${CHART_TGZ}" ] && [ -f "${CHART_TGZ}" ]; then
        CHART_PATH="${CHART_TGZ}"
        echo "Found packaged chart: ${CHART_PATH}"
    elif [ -d "../../../helm/${CHART_NAME}" ]; then
        CHART_PATH="../../../helm/${CHART_NAME}"
        echo "Using chart directory: ${CHART_PATH}"
    else
        echo "ERROR: Chart not found"
        echo "Please run: make helm-package"
        echo "Or ensure chart exists at: ../../../helm/${CHART_NAME}"
        exit 1
    fi
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Installing Helm Chart (Full Install, Not Individual YAMLs)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Chart: ${CHART_NAME}"
echo "Release: ${RELEASE_NAME}"
echo "Namespace: ${NAMESPACE}"
echo "Chart Path: ${CHART_PATH}"
echo "Image: ${IMAGE_REPO}:${IMAGE_TAG}"
echo "Controller Namespace: ${CONTROLLER_NAMESPACE}"
echo "GVKs: ${GVKS}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Verify chart exists and is valid
if [ -f "${CHART_PATH}" ]; then
    echo "✓ Using packaged Helm chart: ${CHART_PATH}"
    if ! helm show chart "${CHART_PATH}" > /dev/null 2>&1; then
        echo "ERROR: Invalid Helm chart package: ${CHART_PATH}"
        exit 1
    fi
elif [ -d "${CHART_PATH}" ]; then
    echo "✓ Using Helm chart directory: ${CHART_PATH}"
    if [ ! -f "${CHART_PATH}/Chart.yaml" ]; then
        echo "ERROR: Invalid Helm chart directory (no Chart.yaml): ${CHART_PATH}"
        exit 1
    fi
else
    echo "ERROR: Chart path does not exist: ${CHART_PATH}"
    exit 1
fi

# Create namespace if it doesn't exist
echo "Creating namespace: ${NAMESPACE}"
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f - || true

# Install Helm chart (FULL INSTALL - all resources via Helm)
echo ""
echo "Installing Helm chart with helm install..."
echo "This will install ALL resources defined in the Helm chart:"
echo "  - Deployment"
echo "  - ServiceAccount"
echo "  - ClusterRole"
echo "  - ClusterRoleBinding"
echo "  - ConfigMap (if configured)"
echo "  - NetworkPolicy (if configured)"
echo ""

helm upgrade --install "${RELEASE_NAME}" "${CHART_PATH}" \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --set image.repository="${IMAGE_REPO}" \
  --set image.tag="${IMAGE_TAG}" \
  --set controller.namespace="${CONTROLLER_NAMESPACE}" \
  --set controller.gvks="${GVKS}" \
  --wait \
  --timeout 5m

echo ""
echo "✅ Helm chart installed successfully!"
echo "Waiting for deployment to be ready..."

# Wait for deployment to be ready
kubectl wait --for=condition=available \
  --timeout=300s \
  deployment/${RELEASE_NAME} \
  -n "${NAMESPACE}" || {
    echo "ERROR: Deployment not ready after 5 minutes"
    echo "Deployment status:"
    kubectl describe deployment "${RELEASE_NAME}" -n "${NAMESPACE}"
    echo ""
    echo "Pod status:"
    kubectl get pods -n "${NAMESPACE}" -l app="${RELEASE_NAME}"
    echo ""
    echo "Pod logs:"
    kubectl logs -n "${NAMESPACE}" -l app="${RELEASE_NAME}" --tail=50
    exit 1
}

echo ""
echo "✅ Deployment is ready!"
echo "Listing installed resources:"
kubectl get all -n "${NAMESPACE}" -l app="${RELEASE_NAME}" || true
kubectl get serviceaccount,configmap -n "${NAMESPACE}" -l app="${RELEASE_NAME}" || true
kubectl get clusterrole,clusterrolebinding | grep "${RELEASE_NAME}" || true
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Helm Chart Installation Complete"
echo "KUTTL will now verify all resources in 01-assert-resources.yaml"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
