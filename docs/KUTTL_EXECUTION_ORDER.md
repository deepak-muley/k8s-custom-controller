# KUTTL Test Execution Order

## How KUTTL Executes Tests

KUTTL executes test files in **alphabetical/numeric order** within each directory level. Files in `steps/` subdirectories run **AFTER** all files in the parent directory.

## Execution Order for `tests/helm-deployment/`

### Phase 1: Main Directory Files (Run First)

KUTTL processes files in the main `tests/helm-deployment/` directory first, in numeric order:

```
1. 00-install.yaml
   └─> Creates namespace: kuttl-test-helm

2. 00-install-helm.sh  (if present, executed automatically)
   └─> Installs Helm chart via "helm install"
   └─> Installs ALL resources: Deployment, SA, RBAC, ConfigMap, etc.

3. 01-assert-resources.yaml
   └─> Asserts all Helm chart resources exist and are ready
   └─> Waits for Deployment (readyReplicas: 1)
   └─> Verifies ServiceAccount exists
   └─> Verifies ClusterRole exists
   └─> Verifies ClusterRoleBinding exists
   └─> Verifies ConfigMap exists
```

**Key Point:** KUTTL waits for assertions in `01-assert-resources.yaml` to pass before moving to `steps/`.

### Phase 2: Steps Directory Files (Run After Main Directory)

**Only after all files in the main directory have completed**, KUTTL processes files in the `steps/` subdirectory, also in numeric order:

```
4. steps/01-verify-controller-functionality.yaml
   └─> Creates test resources (test deployments)
   └─> These will be counted by the controller

5. steps/02-assert-deployments-ready.yaml
   └─> Asserts test deployments are ready (readyReplicas: 1)
   └─> KUTTL waits for this state before continuing

6. steps/03-verify-controller-counting.yaml
   └─> Verifies controller is running
   └─> Verifies controller is counting resources (via logs/metrics)
```

## Complete Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│  KUTTL Test Execution Flow for tests/helm-deployment/       │
└─────────────────────────────────────────────────────────────┘

Step 1: 00-install.yaml
├─> Creates namespace: kuttl-test-helm
└─> ✅ Complete

Step 2: 00-install-helm.sh (script)
├─> Executes: helm install k8s-custom-controller ...
├─> Installs full Helm chart (all resources)
├─> Waits for deployment to be ready
└─> ✅ Complete

Step 3: 01-assert-resources.yaml (ASSERTIONS)
├─> Waits for Deployment (readyReplicas: 1) ⏳
├─> Verifies ServiceAccount exists ⏳
├─> Verifies ClusterRole exists ⏳
├─> Verifies ClusterRoleBinding exists ⏳
├─> Verifies ConfigMap exists ⏳
└─> ✅ All assertions pass

┌─────────────────────────────────────────────────────────────┐
│  ⬇️  Now KUTTL moves to steps/ directory                    │
└─────────────────────────────────────────────────────────────┘

Step 4: steps/01-verify-controller-functionality.yaml
├─> Creates test-deployment-1
├─> Creates test-deployment-2
└─> ✅ Resources created

Step 5: steps/02-assert-deployments-ready.yaml (ASSERTIONS)
├─> Waits for test-deployment-1 (readyReplicas: 1) ⏳
├─> Waits for test-deployment-2 (readyReplicas: 1) ⏳
└─> ✅ All assertions pass

Step 6: steps/03-verify-controller-counting.yaml
├─> Verifies controller pod is running ⏳
├─> Verifies controller is counting resources (logs/metrics) ⏳
└─> ✅ Verification complete

┌─────────────────────────────────────────────────────────────┐
│  ✅ All tests passed                                         │
└─────────────────────────────────────────────────────────────┘
```

## Important Rules

### 1. Files in Main Directory Run First

All files in `tests/helm-deployment/` (not in `steps/`) execute **before** any files in `steps/`.

**Execution Order:**
```
tests/helm-deployment/
├── 00-install.yaml          ← Runs FIRST (step 1)
├── 00-install-helm.sh       ← Runs SECOND (step 2)
├── 01-assert-resources.yaml ← Runs THIRD (step 3)
│
└── steps/                    ← Runs AFTER main directory
    ├── 01-verify-controller-functionality.yaml  ← Runs FOURTH (step 4)
    ├── 02-assert-deployments-ready.yaml         ← Runs FIFTH (step 5)
    └── 03-verify-controller-counting.yaml       ← Runs SIXTH (step 6)
```

### 2. Steps Directory Runs After Main Directory

The `steps/` subdirectory is processed **only after** all files in the parent directory complete successfully.

**Why this matters:**
- Main directory files set up the test environment (Helm chart installation)
- Assertion files in main directory verify the setup worked
- Steps directory files test functionality **after** the setup is verified

### 3. Numeric Prefixes Control Order

Within each directory level, files are sorted by their numeric prefixes:

```
00-*.yaml     ← Runs first
01-*.yaml     ← Runs second
02-*.yaml     ← Runs third
...
10-*.yaml     ← Runs after 09-*.yaml
```

**Example:**
```
00-install.yaml                    ← Runs before 01-assert-resources.yaml
01-assert-resources.yaml           ← Runs before 02-verify-something.yaml
steps/01-verify-functionality.yaml ← Runs before steps/02-assert-ready.yaml
```

### 4. Assertion Files Wait for State

Files with assertions (like `01-assert-resources.yaml`, `02-assert-deployments-ready.yaml`) will **wait** for the resources to match the expected state before continuing.

**From KUTTL Documentation:**
- Assertion files wait until resources match the expected state
- Timeout is configurable (default: 30 seconds per assertion)
- Test fails if assertions don't pass within timeout

## Real Example: Your Test Flow

### Current Structure

```
tests/helm-deployment/
├── 00-install.yaml              ← Step 1: Create namespace
├── 00-install-helm.sh           ← Step 2: Install Helm chart (script)
├── 01-assert-resources.yaml     ← Step 3: Assert Helm resources exist
│
└── steps/                        ← Step 4+: Run after main directory
    ├── 01-verify-controller-functionality.yaml  ← Step 4: Create test resources
    ├── 02-assert-deployments-ready.yaml         ← Step 5: Assert test resources ready
    └── 03-verify-controller-counting.yaml       ← Step 6: Verify controller counting
```

### Execution Timeline

```
Time  Action
────  ──────────────────────────────────────────────────────────────
0s    ✅ 00-install.yaml
      └─> Namespace created

1s    ✅ 00-install-helm.sh
      └─> Helm chart installing...
      └─> Waiting for deployment...

30s   ✅ 01-assert-resources.yaml
      └─> Deployment ready ✅
      └─> ServiceAccount exists ✅
      └─> RBAC exists ✅
      └─> ConfigMap exists ✅

31s   ⬇️  Moving to steps/ directory

31s   ✅ steps/01-verify-controller-functionality.yaml
      └─> Creating test-deployment-1...
      └─> Creating test-deployment-2...

45s   ✅ steps/02-assert-deployments-ready.yaml
      └─> Waiting for test-deployment-1 (readyReplicas: 1) ⏳
      └─> Waiting for test-deployment-2 (readyReplicas: 1) ⏳
      └─> Both deployments ready ✅

60s   ✅ steps/03-verify-controller-counting.yaml
      └─> Verifying controller pod is running ✅
      └─> Checking controller logs for counting... ✅
      └─> Controller is counting resources ✅

61s   ✅ All tests passed!
```

## Why This Order Matters

### Setup Before Tests

1. **Main directory sets up environment:**
   - Creates namespace
   - Installs Helm chart
   - Verifies Helm chart resources exist

2. **Steps directory tests functionality:**
   - Creates test resources (only after Helm chart is installed)
   - Verifies controller functionality (only after controller is running)
   - Tests resource counting (only after test resources exist)

### Assertions Before Next Step

Each assertion file waits for resources to be in the expected state before the next step runs. This ensures:
- ✅ Helm chart is fully deployed before creating test resources
- ✅ Test resources are ready before verifying controller counting
- ✅ Controller is running before checking logs/metrics

## Controlling Execution Order

### Option 1: Use Numeric Prefixes (Current Approach)

```
00-setup.yaml           ← Runs first
01-install.yaml         ← Runs second
02-assert-installed.yaml ← Runs third
steps/
  00-create-test.yaml   ← Runs fourth (after all main directory files)
  01-assert-test.yaml   ← Runs fifth
```

### Option 2: Combine Steps (Alternative)

You can put all steps in the main directory:

```
00-setup.yaml
01-install.yaml
02-assert-installed.yaml
03-create-test.yaml      ← Instead of steps/00-create-test.yaml
04-assert-test.yaml      ← Instead of steps/01-assert-test.yaml
```

**Use `steps/` when:**
- You want to logically group related test steps
- You want clear separation between setup and testing
- You have many steps and want better organization

**Use main directory when:**
- You have a simple linear flow
- You don't need logical grouping
- You prefer a flat structure

## Your Current Test Structure

### What Runs and When

```
✅ Phase 1: Setup (Main Directory)
   1. 00-install.yaml → Creates namespace
   2. 00-install-helm.sh → Installs Helm chart
   3. 01-assert-resources.yaml → Verifies Helm chart resources

✅ Phase 2: Testing (Steps Directory)
   4. steps/01-verify-controller-functionality.yaml → Creates test resources
   5. steps/02-assert-deployments-ready.yaml → Verifies test resources ready
   6. steps/03-verify-controller-counting.yaml → Verifies controller counting
```

## Summary

**When do `tests/helm-deployment/steps/` run?**

✅ **After** all files in `tests/helm-deployment/` (main directory) complete successfully.

**Execution Order:**
1. `00-install.yaml` (main directory)
2. `00-install-helm.sh` (main directory)
3. `01-assert-resources.yaml` (main directory) ← Waits for assertions to pass
4. `steps/01-verify-controller-functionality.yaml` ← Now steps/ runs
5. `steps/02-assert-deployments-ready.yaml` ← In numeric order
6. `steps/03-verify-controller-counting.yaml` ← Last step

**Key Points:**
- Main directory files run first (setup)
- Steps directory runs after (testing)
- Assertion files wait for state before continuing
- Numeric prefixes control order within each directory

## Quick Answer

**Q: When do `tests/helm-deployment/steps/` run?**

**A: After all files in `tests/helm-deployment/` (main directory) complete successfully.**

**Execution Order:**
1. `00-install.yaml` (main directory) - Creates namespace
2. `00-install-helm.sh` (main directory) - Installs Helm chart
3. `01-assert-resources.yaml` (main directory) - Asserts Helm resources exist
4. `steps/01-verify-controller-functionality.yaml` ← **Steps start here**
5. `steps/02-assert-deployments-ready.yaml` - Asserts test resources ready
6. `steps/03-verify-controller-counting.yaml` - Verifies controller counting

**Key Rule:** Main directory files run first (setup), then `steps/` directory runs (testing).

## References

- [KUTTL Documentation](https://kuttl.dev/) - Official documentation
- [KUTTL Test Structure](https://kuttl.dev/docs/test-structure/) - Test execution order
- [KUTTL Usage](./KUTTL_USAGE.md) - Detailed usage guide (includes Helm chart installation)
- [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md) - Quick decision guide
- [E2E Testing Comparison](./E2E_TESTING_COMPARISON.md) - KUTTL vs other E2E approaches
