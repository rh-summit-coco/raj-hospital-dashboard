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

### 🔄 Tekton Pipeline
Automated CI/CD pipeline includes:
- **S2I Build**: Nginx container build from source
- **Security Scanning**: Container vulnerability assessment
- **Automated Testing**: Dashboard functionality validation
- **GitHub Webhooks**: Triggered on code commits
- **Multi-environment**: Dev → Staging → Production

### 🚀 ArgoCD GitOps
GitOps deployment strategy:
- **Dev Environment**: Auto-sync from `main` branch
- **Staging Environment**: Auto-sync with validation
- **Production Environment**: Manual approval + tagged releases
- **Drift Detection**: Automatic remediation
- **Rollback Capability**: One-click previous versions

### 📊 URLs
- **Dashboard**: http://raj-dashboard-raj-compliance-dashboard.apps.uhfgfgde.eastus.aroapp.io
- **ArgoCD**: https://openshift-gitops-server-openshift-gitops.apps.uhfgfgde.eastus.aroapp.io
- **Tekton Console**: OpenShift Console → Pipelines

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