#!/bin/bash

QUAY_USER=quay.io/eesposit

git clone https://github.com/rh-summit-coco/attestation-collector.git
cd attestation-collector

podman build -f Dockerfile.sidecar -t attestation-collector-secure:latest
podman tag attestation-collector-secure:latest $QUAY_USER/attestation-collector-secure:latest
podman push $QUAY_USER/attestation-collector-secure:latest