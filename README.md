# Raj's Hospital Compliance Dashboard

A security compliance dashboard for monitoring confidential computing attestation events in a hospital environment. This dashboard provides real-time visibility into the two-gate security model protecting sensitive AI models and patient data.

## Overview

This dashboard is designed for **Raj**, the Operational Security persona in the Red Hat Summit Confidential Computing demo. It demonstrates how security policies automatically protect hospital assets through:

- **Gate 1**: Code integrity verification via container signatures
- **Gate 2**: Hardware attestation via Trusted Execution Environments (TEEs)

## Features

### 🏥 Hospital-Branded Design
- Professional medical color scheme (deep blues, teals, medical greens)
- Clean, accessible interface suitable for compliance officers
- Responsive design that works on various screen sizes

### 🛡️ Security Monitoring
- **Real-time status**: Visual indicators showing overall system compliance
- **Two-gate monitoring**: Separate status for code integrity and TEE attestation
- **Event logging**: Chronological view of attestation successes and failures
- **Incident alerts**: Prominent alerts when security violations are detected

### 📊 Demo Scenarios
- **Normal Operations**: Shows successful attestation flow (green status)
- **Security Incident**: Shows failed attestation attempt (red alert)
- Toggle between scenarios for demo purposes

## Architecture

### Phase 1: Static Demo Dashboard
- Self-contained HTML/CSS/JavaScript application
- Mock data for demo scenarios
- Manual scenario switching via UI buttons

### Phase 2: Live Integration (Future)
- Integration with Pradipta's sidecar API
- Real-time event streaming from Trustee KBS
- WebSocket updates for immediate incident notification

### Attestation Collector (source of truth for reports)

The **dashboard backend** in this repo calls the Attestation Collector’s `GET /api/v1/reports` and mirrors what it returns. It does not own long-term storage of attestation data.

The **Attestation Collector** is built from a separate repo: [`rh-summit-coco/attestation-collector`](https://github.com/rh-summit-coco/attestation-collector) (see `buildconfig/buildconfig-collector.yaml`). It receives reports from sidecars and exposes them to the dashboard. **If the collector keeps old reports** (e.g. in-memory map or cache) **and keeps listing them** when sidecars stop sending updates, **the dashboard will keep showing them** until the collector stops returning them.

**To fix stale reports at the source:** implement retention / eviction in the attestation-collector (e.g. drop entries when no new report arrives for a pod within a TTL, remove reports for deleted pods, or cap list age). The dashboard can apply extra guards (env vars like `CACHE_MAX_WORKLOAD_AGE_SECONDS`, cache clearing on failed fetches, fingerprint-based timestamp handling) but the authoritative fix is in the collector.

## Deployment

### Local Development
```bash
# Serve locally for testing
python3 -m http.server 8080
# Open http://localhost:8080
```

### OpenShift Deployment
```bash
# Apply Kubernetes manifests
oc apply -f kubernetes/deployment.yaml

# Check deployment
oc get pods -n trustee-operator-system -l app=raj-hospital-dashboard

# Get dashboard URL
oc get route raj-dashboard-route -n trustee-operator-system
```

### Container Build
```bash
# Build container image
podman build -t raj-hospital-dashboard:latest .

# Tag for quay.io
podman tag raj-hospital-dashboard:latest quay.io/rh-summit-cooc/raj-hospital-dashboard:latest

# Push to registry
podman push quay.io/rh-summit-cooc/raj-hospital-dashboard:latest
```

## CI/CD Architecture

### Secure CI Pipeline

The RAJ Dashboard uses a Tekton-based secure CI pipeline that implements supply chain security.

#### Pipeline Stages

```
┌─────────────┐   ┌──────────────┐   ┌─────────────────┐   ┌─────────────────┐
│ clone-source│ → │ run-go-tests │ → │ security-scan   │ → │ build-container │
└─────────────┘   └──────────────┘   └─────────────────┘   └─────────────────┘
                                                                    │
┌─────────────┐   ┌──────────────┐   ┌─────────────────┐            │
│   deploy    │ ← │ generate-sbom│ ← │   sign-image    │ ← ─────────┘
└─────────────┘   └──────────────┘   └─────────────────┘
       │
       ▼
┌─────────────────┐
│integration-tests│
└─────────────────┘
```

| Stage | Tool | Description |
|-------|------|-------------|
| **clone-source** | git-clone | Clone repository from GitHub |
| **run-go-tests** | go test | Run unit tests with coverage |
| **security-scan** | Custom | Scan for hardcoded secrets, unsafe imports |
| **build-container** | buildah | Build OCI container (amd64) |
| **vulnerability-scan** | Clair | Scan for CVEs (blocks on CRITICAL) |
| **sign-image** | Cosign | Keyless cryptographic signing |
| **generate-sbom** | Custom | Generate Software Bill of Materials |
| **deploy** | oc | Deploy to OpenShift namespace |
| **integration-tests** | curl | Test API endpoints post-deploy |

#### Deploy Pipeline

```bash
# Apply tasks first
oc apply -f tekton/tasks.yaml -n raj-compliance-dashboard

# Apply pipeline
oc apply -f tekton/secure-ci-pipeline.yaml -n raj-compliance-dashboard

# Run pipeline manually
oc create -f - <<EOF
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: raj-dashboard-run-
  namespace: raj-compliance-dashboard
spec:
  pipelineRef:
    name: raj-dashboard-secure-pipeline
  workspaces:
  - name: source-ws
    volumeClaimTemplate:
      spec:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: 1Gi
EOF
```

#### Run Tests Locally

```bash
cd backend
go test -v -cover ./...
```

### ArgoCD GitOps

GitOps deployment via ArgoCD:
- **Source**: `hospital-demo-gitops` repository
- **Auto-sync**: Enabled for dev environment
- **Drift Detection**: Automatic remediation

### URLs

| Service | URL |
|---------|-----|
| Dashboard | https://raj-hospital-dashboard-raj-compliance-dashboard.apps.uhfgfgde.eastus.aroapp.io |
| Attestation API | https://attestation-collector-raj-compliance-dashboard.apps.uhfgfgde.eastus.aroapp.io |
| ArgoCD | https://openshift-gitops-server-openshift-gitops.apps.uhfgfgde.eastus.aroapp.io |
| Tekton | OpenShift Console → Pipelines |

## Demo Usage

### Scenario 1: Normal Operations
1. Click "✅ Normal Operations" button
2. Shows green status with successful attestation events
3. Both gates show "PASSING" status
4. Event log shows recent successful operations

### Scenario 2: Security Incident
1. Click "🚨 Security Incident" button
2. Shows red alert with security violation details
3. Gate 1 passes (signed container) but Gate 2 fails (attestation)
4. Clear incident details: timestamp, workload, reason, action taken

### CI/CD Demo Flow
1. **Make a code change** and commit to GitHub
2. **Tekton pipeline** automatically triggers
3. **Build completes** and deploys to development
4. **ArgoCD syncs** changes to staging environment
5. **Manual promotion** to production after validation

## Demo Story Integration

This dashboard supports the "Day in the Life of Raj" demo flow:

1. **Raj's Hero Moment**: Dashboard visually proves his policies work
2. **Attack Scenario**: Shows real-time blocking of tampered workload
3. **Zero Trust Success**: Demonstrates automatic policy enforcement
4. **Compliance Evidence**: Provides audit trail for security incidents

## Technical Details

- **Frontend**: Pure HTML5, CSS3, JavaScript (no external dependencies)
- **Styling**: CSS Grid, Flexbox, CSS animations
- **Container**: Red Hat UBI 9 with nginx-120
- **Platform**: OpenShift 4.x compatible
- **Security**: Non-root container, minimal attack surface

## Color Scheme

```css
--hospital-primary: #1e4a72;      /* Deep medical blue */
--hospital-secondary: #2e8b87;    /* Teal green */
--hospital-success: #28a745;      /* Medical green */
--hospital-danger: #dc3545;       /* Emergency red */
--hospital-accent: #f8f9fa;       /* Clean white */
```

## Future Enhancements

- [ ] Live data integration with Trustee KBS logs
- [ ] WebSocket real-time updates
- [ ] Historical trend analysis
- [ ] Configurable alert thresholds
- [ ] Multi-cluster monitoring
- [ ] Export compliance reports
- [ ] Mobile app companion

## License

Part of the Red Hat Summit Confidential Computing demonstration materials.