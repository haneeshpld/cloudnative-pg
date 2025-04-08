#!/bin/bash

# Set variables
NAMESPACE="cnpg-system"
BUNDLE_IMAGE="ghcr.io/haneeshpld/cloudnative-pg-testing:main-bundle"
VERSION="1.0.0"

# Function to check command success
check_success() {
    if [ $? -ne 0 ]; then
        echo "❌ Error: $1"
        exit 1
    fi
}

# Ensure required commands are available
command -v oc >/dev/null 2>&1 || { echo "❌ Error: OpenShift CLI (oc) not found!"; exit 1; }
command -v make >/dev/null 2>&1 || { echo "❌ Error: make command not found!"; exit 1; }
command -v operator-sdk >/dev/null 2>&1 || { echo "❌ Error: operator-sdk not found!"; exit 1; }

unset DOCKER_HOST

# Step 1: Check if logged into OpenShift
oc whoami >/dev/null 2>&1
check_success "Not logged into OpenShift. Please log in using 'oc login'"

# Step 2: Delete the existing namespace
echo "🚀 Deleting existing namespace: $NAMESPACE"
oc delete namespace $NAMESPACE --ignore-not-found=true
check_success "Failed to delete namespace"

# Step 3: Wait for namespace deletion to complete
while oc get namespace $NAMESPACE >/dev/null 2>&1; do
    echo "⏳ Waiting for namespace $NAMESPACE to be fully deleted..."
    sleep 5
done
echo "✅ Namespace deleted successfully."

# Step 4: Recreate the namespace
echo "🚀 Creating namespace: $NAMESPACE"
oc create namespace $NAMESPACE
check_success "Failed to create namespace"

oc adm policy add-scc-to-user anyuid -z default -n $NAMESPACE

# Step 5: Build the OLM catalog
echo "🚀 Building OLM catalog with version: $VERSION"
make olm-catalog VERSION=$VERSION
check_success "Failed to build OLM catalog"

# Step 6: Run the operator bundle
echo "🚀 Running operator bundle: $BUNDLE_IMAGE"
operator-sdk run bundle $BUNDLE_IMAGE --timeout=600s
check_success "Failed to run the operator bundle"

echo "✅ Deployment completed successfully!"
