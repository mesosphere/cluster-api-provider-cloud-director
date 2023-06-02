include Makefile

# The upstream 'release-manifests' target does not correctly override the image.
# We work around this by using `kustomize edit set image`.
release-manifests: kustomize
	mkdir -p $(MANIFEST_DIR)
	cd config/manager && $(KUSTOMIZE) edit set image projects.registry.vmware.com/vmware-cloud-director/cluster-api-provider-cloud-director=$(REGISTRY)/$(CAPVCD_IMG):$(VERSION)
	$(KUSTOMIZE) build config/default > $(MANIFEST_DIR)/infrastructure-components.yaml
