# ArgoCD GitOps Deployment for Raj's Dashboard

This directory contains ArgoCD Application manifests for GitOps deployment of the hospital compliance dashboard.

## Architecture

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│     Development     │    │      Staging        │    │    Production       │
│  raj-compliance-    │    │ raj-compliance-     │    │ raj-compliance-     │
│     dashboard       │────│   dashboard-        │────│   dashboard-prod    │
│                     │    │     staging         │    │                     │
│ Auto-sync: ✅       │    │ Auto-sync: ✅        │    │ Manual-sync: 👤     │
│ Source: main branch │    │ Source: main branch │    │ Source: v* tags     │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
```

## Applications

### 1. Development Environment
- **Name**: `raj-hospital-dashboard`
- **Namespace**: `raj-compliance-dashboard`
- **Source**: `main` branch
- **Sync**: Automated with self-healing
- **Purpose**: Active development and testing

### 2. Staging Environment
- **Name**: `raj-dashboard-staging`
- **Namespace**: `raj-compliance-dashboard-staging`
- **Source**: `main` branch
- **Sync**: Automated
- **Purpose**: Pre-production validation

### 3. Production Environment
- **Name**: `raj-dashboard-production`
- **Namespace**: `raj-compliance-dashboard-prod`
- **Source**: Git tags (v1.0.0, v1.1.0, etc.)
- **Sync**: Manual approval required
- **Purpose**: Live production environment

## Project Configuration

The `confidential-computing` AppProject provides:
- **Repository Access**: `rh-summit-coco/*` repositories
- **Namespace Access**: All dashboard namespaces + trustee-operator-system
- **Resource Permissions**: Deployments, Services, Routes, BuildConfigs
- **RBAC**: Admin group access

## Deployment Workflow

1. **Developer commits** to `main` branch
2. **Tekton Pipeline** builds and tests
3. **ArgoCD syncs** to development environment
4. **ArgoCD syncs** to staging environment
5. **Manual verification** in staging
6. **Git tag** created for production release
7. **Manual sync** to production

## Installation

### Prerequisites
- ArgoCD installed on cluster
- Access to `rh-summit-coco` GitHub organization

### Deploy Applications
```bash
# Apply the project and applications
oc apply -f argocd/application.yaml

# Check application status
oc get applications -n argocd

# View application details
argocd app get raj-hospital-dashboard
```

### Monitor Deployments
```bash
# Watch ArgoCD UI
https://argocd-server-argocd.apps.uhfgfgde.eastus.aroapp.io

# CLI monitoring
argocd app sync raj-hospital-dashboard
argocd app wait raj-hospital-dashboard --health
```

## GitOps Benefits for Raj's Dashboard

### 🔐 Security & Compliance
- **Audit trail**: All changes tracked in Git
- **Policy enforcement**: ArgoCD validates deployments
- **Least privilege**: Namespace-scoped permissions
- **Automated rollbacks**: Failed deployments auto-revert

### 📊 Operational Excellence
- **Declarative config**: Infrastructure as code
- **Multi-environment**: Dev → Staging → Production
- **Automated sync**: Reduces manual errors
- **Self-healing**: Drift detection and correction

### 🚀 Developer Experience
- **Git-based workflow**: Standard development process
- **Preview deployments**: Test changes safely
- **Easy rollbacks**: One-click previous versions
- **Status visibility**: Clear deployment state

This GitOps setup aligns perfectly with Raj's security compliance requirements, providing full traceability and automated policy enforcement for the confidential computing dashboard.