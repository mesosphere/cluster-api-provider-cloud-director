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
		--file d2iq.Dockerfile \
		.

.PHONY: docker-buildx-builder
docker-buildx-builder:
	docker buildx inspect --bootstrap capvcd &>/dev/null || docker buildx create --name capvcd

# The upstream 'release-manifests' target does not correctly override the image.
# We work around this by using `kustomize edit set image`.
release-manifests: $(KUSTOMIZE)
	mkdir -p templates && \
	cd config/manager && \
		$(GITROOT)/$(KUSTOMIZE) edit set image projects.registry.vmware.com/vmware-cloud-director/cluster-api-provider-cloud-director=$(REGISTRY)/$(CAPVCD_IMG):$(VERSION)
	$(GITROOT)/$(KUSTOMIZE) build config/default > templates/infrastructure-components.yaml

.PHONY: build-within-docker
build-within-docker: vendor
	mkdir -p /build/cluster-api-provider-cloud-director
	CGO_ENABLED=0 go build -ldflags "-X github.com/vmware/$(CAPVCD_IMG)/version.Version=${VERSION}" -o /build/vcloud/cluster-api-provider-cloud-director main.go
