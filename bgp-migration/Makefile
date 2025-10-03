CILIUM_HELM_REPO ?= oci://quay.io/cilium-charts-dev/cilium
CILIUM_VERSION ?= 1.16.12
COIL_VERSION ?= 2.8.0
CKE_VERSION ?= 1.31.0
BIRD_VERSION ?= 3.0.0
BGP_VARIANT ?= metallb

ifeq ($(BGP_VARIANT),metallb)
  CILIUM_DIR = metallb/base
else
  CILIUM_DIR = bgp-cp/base
endif

ifdef PATCHED
  CILIUM_DIR := $(CILIUM_DIR)/../overlay/patched
endif

.PHONY: deploy
deploy: update-coil update-rivers
	# must start containerlab in the background before kind to utilize its retries.
	sudo containerlab -t topo.yaml deploy &

	kind create cluster --config cluster.yaml
	kind load docker-image --name bgp-migration ghcr.io/cybozu-go/coil:2.8.0
	kind load docker-image --name bgp-migration cybozu-go/lbipam-annotate:latest

	# remove taint from control plane node
	kubectl taint nodes bgp-migration-control-plane node-role.kubernetes.io/control-plane:NoSchedule-

	$(MAKE) install-coil
	$(MAKE) install-cilium;

	kubectl apply -f configs/coil_default_pool.yaml

	# create pods for testing
	kubectl apply -f test_pod.yaml
	kubectl apply -f test_lb.yaml

	# Add annotations to LB services
	# kubectl apply -k lbipam-annotate/config/apply

.PHONY: install-coil
install-coil:
	# install coil
	kubectl apply -k coil/base
	kubectl wait --for=condition=Ready -n kube-system pod -l app.kubernetes.io/name=coil --timeout=300s

.PHONY: install-cilium
install-cilium:
ifdef PATCHED
	kind load docker-image --name bgp-migration localhost:5000/cilium/cilium-dev:local
	kind load docker-image --name bgp-migration localhost:5000/cilium/operator-generic:local
endif

	# add configmap for bgp
	if [ "$(BGP_VARIANT)" = "metallb" ]; then \
		kubectl apply -f metallb/bgp-config.yaml; \
	fi

	# install cilium
	kubectl apply -k $(CILIUM_DIR)

	# kubectl wait behaves weirdly when ran right after applying manifests for some reason
	sleep 5

	kubectl wait --for=condition=Ready -n kube-system pod -l app.kubernetes.io/part-of=cilium --timeout=300s
	cilium status --namespace kube-system --ignore-warnings

.PHONY: apply-bgp-cp
apply-bgp-cp:
	# apply bgp manifests
	kubectl apply -f bgp-cp/bgp_lb.yaml

.PHONY: install-bgp-cp
install-bgp-cp:
	$(MAKE) install-cilium BGP_VARIANT=bgp-cp

.PHONY: patch-bgp-config
patch-bgp-config:
	kubectl apply -f metallb/bgp-config-no-pool.yaml


.PHONY: apply-baseline-lbipam
apply-baseline-lbipam:
	kubectl apply -f metallb/_tmp_baseline_pool.yaml

.PHONY: apply-lbipam
apply-lbipam:
	kubectl apply -f metallb/lbipam.yaml

.PHONY: remove-annotation
remove-annotation:
	kubectl apply -k lbipam-annotate/config/remove

.PHONY: destroy
destroy:
	sudo containerlab -t topo.yaml destroy -c
	kind delete clusters bgp-migration
	rm -f .topo.yaml.bak

.PHONY: reload
reload:
	$(MAKE) destroy
	$(MAKE) deploy

update-coil: coil/base/coil.yaml
coil/base/coil.yaml:
	rm -rf /tmp/work-coil
	mkdir -p /tmp/work-coil
	curl -sSfL https://github.com/cybozu-go/coil/archive/v$(COIL_VERSION).tar.gz | tar -C /tmp/work-coil -xzf - --strip-components=1
	cd /tmp/work-coil/v2 ; sed -i -E 's,^#(- config/pod/compat_calico.yaml),\1,' kustomization.yaml
	cd /tmp/work-coil/v2 ; $(MAKE) certs
	cp configs/netconf.json /tmp/work-coil/v2/netconf.json
	kustomize build /tmp/work-coil/v2 > coil/base/coil.yaml
	rm -rf /tmp/work-coil

update-cilium:
	rm -rf /tmp/work-cilium
	mkdir -p /tmp/work-cilium
	git clone --depth 1 --branch v$(shell echo $(CILIUM_VERSION) | cut -d \. -f 1,2,3) https://github.com/cilium/cilium /tmp/work-cilium
	cd /tmp/work-cilium
	helm template /tmp/work-cilium/install/kubernetes/cilium/ \
		--namespace=kube-system \
		--values $(CILIUM_DIR)/values.yaml > $(CILIUM_DIR)/base/upstream.yaml
	rm -rf /tmp/work-cilium

update-rivers: tools/rivers
tools/rivers:
	rm -rf /tmp/work-cke
	mkdir -p /tmp/work-cke
	curl -sSfL https://github.com/cybozu-go/cke/archive/v$(CKE_VERSION).tar.gz | tar -C /tmp/work-cke -xzf - --strip-components=1
	cd /tmp/work-cke/tools ; mkdir bin && GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/rivers ./rivers
	cp /tmp/work-cke/tools/bin/rivers tools/
	rm -rf /tmp/work-cke

build-bird: 
	docker build -f bird.Dockerfile -t bird:$(BIRD_VERSION) --platform linux/arm64 .

build-lbipam-annotate: 
	cd ./lbipam-annotate; docker build -t cybozu-go/lbipam-annotate --platform linux/arm64 .
