# Sandbox for Cilium BGP CP V2 

## Project Structure
- ./service: Contains exact copy from [a Demo on Cilium Repo](https://github.com/cilium/cilium/tree/v1.16.11/contrib/containerlab/bgpv2/service)
- ./bird: Adjust the demo by using Bird instead of FRR. It also makes Cilium agent peer with the Bird instance on the same node that it is on.
- ./bird-coil: Further adjust the demo to use [Coil](https://github.com/cybozu-go/coil) as main CNI. Cilium is used only for L4LB, kube-proxy replacement, and Network Policy.
