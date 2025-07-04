# Cilium BGP Control Plane V2 with Coil and Bird

## Topology
```
                        +------------------------+
                        |        router          |
                        |     AS 65001 (BIRD)    |
                        +-----------+------------+
                                    |
                 +------------------+------------------+
                 |                                     |
     +-----------+-----------+           +-------------+-----------+
     |       server0         |           |       server1           |
     |  AS 65001 (BIRD)      |           |  AS 65001 (BIRD)        |
     | +-------------------+ |           | +---------------------+ |
     | |  Cilium BGP CP    | |           | |  Cilium BGP CP      | |
     | | AS 65002 (lo peer)| |           | |  AS 65002 (lo peer) | |
     | +-------------------+ |           | +---------------------+ |
     +-----------------------+           +-------------------------+
```

## Requirements
- kind
- containerlab
- Go

## Usage
```sh
# Build Bird image used by containerlab nodes
make build-bird

# Deploy
make deploy

# Test if BGP CP is working
# 1. Pod to Pod
# Test connectivity by pinging a pod on server1 from another pod on server0
# Retrieve ip for pod on server1
POD_IP=kubectl get pods ubuntu-worker -oyaml | yq .status.podIP
# Or with fish
set POD_IP (kubectl get pods ubuntu-worker -oyaml | yq .status.podIP)
# Ping the Pod from a pod on another node
kubectl exec ubuntu -it -- ping $POD_IP

# 2. L4LB with GENEVE DSR
# Test connectivity by making http request to a pod behind Service of type: LoadBalancer 
docker exec clab-bgpv2-bird-coil-router0 curl -sSL 20.1.10.2

# Destroy
make destroy

# Reload (Destroy then Deploy)
make reload

```

