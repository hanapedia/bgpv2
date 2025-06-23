CILIUM_HELM_REPO ?= oci://quay.io/cilium-charts-dev/cilium
CILIUM_VERSION ?= 1.16.0-dev-dev.247-main-8c97a2119d
COIL_VERSION ?= 2.8.0

deploy:
	kind create cluster --config cluster.yaml
	kind load docker-image --name bgpv2-cplane-dev-service bird:latest

	sudo containerlab -t topo.yaml deploy
	# remove taint from control plane node
	kubectl taint nodes bgpv2-cplane-dev-service-control-plane node-role.kubernetes.io/control-plane:NoSchedule-

install-with-coil:
	# install coil
	kubectl apply -f configs/coil.yaml
	kubectl wait --for=condition=Ready -n kube-system pod -l app.kubernetes.io/name=coil
	kubectl apply -f configs/coil_default_pool.yaml

	# install cilium
	helm install cilium -n kube-system $(CILIUM_HELM_REPO) --version $(CILIUM_VERSION) -f configs/values.yaml
	cilium status --wait --namespace kube-system

install-cilium:
	# install cilium
	helm install cilium -n kube-system $(CILIUM_HELM_REPO) --version $(CILIUM_VERSION) -f values.yaml
	cilium status --wait --namespace kube-system

destroy:
	sudo containerlab -t topo.yaml destroy -c
	kind delete clusters bgpv2-cplane-dev-service
	rm -f .topo.yaml.bak

reload:
	$(MAKE) destroy
	$(MAKE) deploy

apply-bgp:
	kubectl apply -f bgp.yaml

apply-service:
	kubectl apply -f service.yaml

apply-lb:
	kubectl apply -f lb-ip.yaml

update-coil:
	rm -rf /tmp/work-coil
	mkdir -p /tmp/work-coil
	curl -sSfL https://github.com/cybozu-go/coil/archive/v$(COIL_VERSION).tar.gz | tar -C /tmp/work-coil -xzf - --strip-components=1
	cd /tmp/work-coil/v2 ; sed -i -E 's,^#(- config/pod/compat_calico.yaml),\1,' kustomization.yaml
	cd /tmp/work-coil/v2 ; $(MAKE) certs
	cp configs/netconf.json /tmp/work-coil/v2/netconf.json
	kustomize build /tmp/work-coil/v2 > configs/coil.yaml
	rm -rf /tmp/work-coil
