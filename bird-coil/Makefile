CILIUM_HELM_REPO ?= oci://quay.io/cilium-charts-dev/cilium
CILIUM_VERSION ?= 1.16.0-dev-dev.247-main-8c97a2119d
COIL_VERSION ?= 2.8.0
CKE_VERSION ?= 1.31.0
BIRD_VERSION ?= 2.0.0

deploy: update-coil update-rivers
	# must start containerlab in the background before kind to utilize its retries.
	sudo containerlab -t topo.yaml deploy &

	kind create cluster --config cluster.yaml
	kind load docker-image --name bgpv2-bird-coil bird:$(BIRD_VERSION)
	kind load docker-image --name bgpv2-bird-coil ghcr.io/cybozu-go/coil:2.8.0

	# remove taint from control plane node
	kubectl taint nodes bgpv2-bird-coil-control-plane node-role.kubernetes.io/control-plane:NoSchedule-

	$(MAKE) install-with-coil

	# restart coredns
	kubectl rollout restart -n kube-system deployment coredns

	kubectl apply -f bgp_lb.yaml
	kubectl apply -f lb_test.yaml
	kubectl apply -f test_pod.yaml

install-with-coil:
	# install coil
	kubectl apply -k coil/base
	kubectl wait --for=condition=Ready -n kube-system pod -l app.kubernetes.io/name=coil --timeout=300s


	# install cilium
	helm install cilium -n kube-system $(CILIUM_HELM_REPO) --version $(CILIUM_VERSION) -f values.yaml
	cilium status --wait --namespace kube-system
    # apply default Address pool
	kubectl apply -f configs/coil_default_pool.yaml

destroy:
	sudo containerlab -t topo.yaml destroy -c
	kind delete clusters bgpv2-bird-coil
	rm -f .topo.yaml.bak

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

update-rivers: tools/rivers
tools/rivers:
	rm -rf /tmp/work-cke
	mkdir -p /tmp/work-cke
	curl -sSfL https://github.com/cybozu-go/cke/archive/v$(CKE_VERSION).tar.gz | tar -C /tmp/work-cke -xzf - --strip-components=1
	cd /tmp/work-cke/tools ; mkdir bin && GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/rivers ./rivers
	cp /tmp/work-cke/tools/bin/rivers tools/
	rm -rf /tmp/work-cke

build-bird: 
	docker build -f bird.Dockerfile -t bird:$(BIRD_VERSION) --platform linux/amd64 .
