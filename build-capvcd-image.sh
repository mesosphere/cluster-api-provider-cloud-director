#!/bin/bash

set -euxo pipefail

export VERSION=v1.2.0-d2iq.beta.0
export REGISTRY=docker.io/kaiwalyarjoshi
export CAPVCD_IMG=cluster-api-vcd-controller
export PLATFORMS="linux/amd64,linux/arm64"
#export PLATFORMS="linux/amd64"
#export PLATFORMS="linux/arm64"

make docker-buildx-builder \
            --makefile d2iq.Makefile

make docker-build-capvcd \
  PLATFORM="${PLATFORMS}" \
  REGISTRY="${REGISTRY}" \
  CAPVCD_IMG="${CAPVCD_IMG}" \
  VERSION="${VERSION}"

make push-capvcd-image \
            --makefile d2iq.Makefile \
            REGISTRY=${REGISTRY} \
            CAPVCD_IMG=${CAPVCD_IMG} \
            VERSION=${VERSION}

