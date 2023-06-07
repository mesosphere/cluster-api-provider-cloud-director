include Makefile

PLATFORMS := "linux/amd64,linux/arm64"

# The multi-platform build must push its output to a registry, because the Docker Engine image store does not support
# multi-platform images.  This will be addressed when Docker Engine supports the containerd image store. See
# https://docs.docker.com/storage/containerd.
.PHONY: push-capvcd-image
push-capvcd-image: docker-buildx-builder generate fmt vet vendor
	docker buildx build \
		--builder capvcd \
		--platform $(PLATFORMS) \
		--output=type=registry \
		--build-arg VERSION=$(VERSION) \
		--tag $(REGISTRY)/$(CAPVCD_IMG):$(VERSION) \
		--file Dockerfile \
		.

.PHONY: docker-buildx-builder
docker-buildx-builder:
	docker buildx inspect --bootstrap capvcd &>/dev/null || docker buildx create --name capvcd

# The upstream 'release-manifests' target does not correctly override the image.
# We work around this by using `kustomize edit set image`.
release-manifests: kustomize
	mkdir -p $(MANIFEST_DIR)
	cd config/manager && $(KUSTOMIZE) edit set image projects.registry.vmware.com/vmware-cloud-director/cluster-api-provider-cloud-director=$(REGISTRY)/$(CAPVCD_IMG):$(VERSION)
	$(KUSTOMIZE) build config/default > $(MANIFEST_DIR)/infrastructure-components.yaml
