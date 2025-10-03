# Steps for Different Upgrade Procedures

## Single Operator Restart
This procedure will apply LBIPAM and BGP CP changes in one operator restart.

### Pre-Steps
1. Deploy Cilium with MetalLB installed.
```sh
make deploy
```
This target will create kind cluster on top of a BGP network.
In this cluster, Coil is installed as the primary CNI and Cilium is installed for L4LB using MetalLB.
MetalLB is configured with two LBIP pools: default (20.1.10.0/24) and bastion (20.1.20.0/24).
LBIPAM is already enabled. It is just in a dormant state, waiting for the first LBIPPool resource to be created.

It will also create 2 Deployments and 4 LoadBalancer Services.
Both pools will each assign LBIPs to 2 LoadBalancers.
Each LoadBalancer Service is also labeled so that LBIPAM can reference it in later steps.

It will also run a one time job that adds "lbipam.cilium.io/ips" annotation with the currently assigned LBIP as the value to each LoadBalancer Service.

### Steps
2. Apply Cilium changes
```sh
make install-bgp-cp PATCHED=1
```
This will update Cilium to disable MetalLB and enable BGPCP.
It will also use patched operator image that includes the option to configure the number of LBIPPools required before reconciling Services to assign LBIP.
The minimum number of LBIPPools required are set to 2 which is same number of pools that MetalLB used.

This will immediately restart the operator but it does not restart the agent as it is configured with OnDelete updateStrategy.
Which means that MetalLB's IPAM feature will be diabled right away, but the BGP speaker feature will not.

3. Create LBIPPools
```sh
make apply-lbipam
```
This will create 2 LBIPPool resources that matche the IP range defined for MetalLB.
This will trigger the reconciliation of the LBIP assigned to existing Services, however, it should not cause any reassignments because the LB Services request the IP from the same range.

4. Create BGP CP Resources
```sh
make apply-bgp-cp
```
This will create `CiliumBGPClusterConfig`, `CiliumBGPPeerConfig`, and `CiliumBGPAdvertisement` but it will not be used until agents are restarted.

5. Restart each agent
```sh
kubectl delete pod -n kube-system cilium-xxx
# leave some time and observe BGP adverts
kubectl delete pod -n kube-system cilium-yyy
# leave some time and observe BGP adverts
kubectl delete pod -n kube-system cilium-zzz
```
Restart agents by deleting the pod one by one. The new pods will run BGP CP instead of MetalLB.

### Useful watch commands to run on the side
Watching for LBIP reassignments.
```sh
kubectl get svc -A --field-selector="spec.type=LoadBalancer" -w
```
Watching for BGP advertisements for LBIPs
```sh
docker logs clab-bgp-migration-router0 -f | grep -e '20.1.10.' -e '20.1.20.'
```
