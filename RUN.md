# How to run

* Build the collector. It runs locally so you can run it via buildconfig.
```
oc apply -f buildconfig/buildconfig-collector.yaml

oc start-build attestation-collector-secure
```

<!-- * Compile backend:
```
podman build -f Dockerfile.backend -t raj-dashboard-backend:latest .

QUAY_USER=quay.io/eesposit
podman tag raj-dashboard-backend:latest $QUAY_USER/raj-dashboard-backend:latest

podman push $QUAY_USER/raj-dashboard-backend:latest
``` -->
* Build the backend. It runs locally so you can run it via buildconfig.
```
oc apply -f buildconfig/buildconfig-backend.yaml

oc start-build raj-dashboard-backend
```

<!-- * Build the sidecar, if you want to run something locally
```
oc apply -f buildconfig/buildconfig-sidecar.yaml

oc start-build attestation-sidecar-secure
``` -->

* Build the sidecar and push it to quay.io: run `buildconfig/build-sidecar.sh`. Essentially this goes into another repo, builds the sidecar container for you and pushes it to quay.

* Create Route and Service for the collector `oc apply -f deploy/network-att-collector.yaml`. This is needed by `setup_certificates.sh`.

* Get the attestation collector route `oc get route attestation-collector-secure -n raj-compliance-dashboard -o jsonpath='https://{.spec.host}'`

* Create the necessary certificates with `./deploy/setup_certificates.sh`. Note that some will be created in the `raj-compliance-dashboard` ns (used by the attestation collector) and some in the `trustee-operator-system` ns (provided to the sidecar attached to the CoCo container). The former secret is needed by `deployment-att.collector.yaml`

* Run the collector `oc apply -f deploy/deployment-att-collector.yaml`

* Run the dashboard `oc apply -f deploy/deployment-dashboard.yaml`

* Get the dashboard route `oc get route raj-dashboard-route -n raj-compliance-dashboard -o jsonpath='https://{.spec.host}'`. You can open this link in your browser.

* Apply the sidecar to the CoCo pod. Remember as `COLLECTOR_URL` you need to use the url obtained for the attestation-collector:
```
spec:
  containers:
  - name: coco-beacon-sidecar
    image: image-registry.openshift-image-registry.svc:5000/raj-compliance-dashboard/attestation-sidecar-secure:latest
    env:
    - name: POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    - name: COLLECTOR_URL
      value: "https://attestation-collector-secure-raj-compliance-dashboard.apps.uml9d8rjadd09bf3f0.eastus.aroapp.io"
    - name: TEE_TYPE
      value: "tdx"
    - name: REPORT_INTERVAL
      value: "30"
    - name: CLIENT_CERT_FILE
      value: "/etc/certs/client.crt"
    - name: CLIENT_KEY_FILE
      value: "/etc/certs/client.key"
    - name: CA_CERT_FILE
      value: "/etc/certs/ca.crt"
    volumeMounts:
    - name: client-certs
      mountPath: /etc/certs
      readOnly: true
  volumes:
  - name: client-certs
    secret:
      secretName: attestation-collector-certs
```

