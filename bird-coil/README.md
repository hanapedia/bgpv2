# Cilium BGP Control Plane V2 with Coil and Bird

## Topology
- Coil runs as the main CNI plugin, responsible for IPAM
- Bird on each node and router advertises Pod IPs added by Coil to its peers via iBGP
- Each Cilium agent peers with Bird running on the same node via lo and advertises LoadBalancerIPs
  - Cilium is configured to use DSR with GENEVE for LoadBalancers with `externalTrafficPolicy: Cluster`.
- ClusterIP routing is handled by Cilium's kube-proxy replacement feature.
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
### Deploying the demo.
```sh
# Build Bird image used by containerlab nodes
make build-bird

# Deploy
make deploy
```

### Test Pod IP connectivity
1. Pod to Pod
Test connectivity by pinging a pod on server1 from another pod on server0.
The demo deploys pods on server0 and server1 by default.
```sh
# Retrieve ip for pod on server1
POD_IP=kubectl get pods ubuntu-worker -oyaml | yq .status.podIP
# Or with fish
set POD_IP (kubectl get pods ubuntu-worker -oyaml | yq .status.podIP)
# Ping the Pod from a pod on another node
kubectl exec ubuntu -it -- ping $POD_IP
```

### Test L4LB with GENEVE DSR
Test connectivity by making http request to a pod behind Service of type: LoadBalancer 
The demo deploys an http server pod on server0 and a LoadBalancer Service with IP of 20.1.10.2
```sh
docker exec clab-bgpv2-bird-coil-router0 curl -sSL 20.1.10.2
```

For testing purposes, there is only one pod deployed on server0, and router uses L3 hash for ECMP.
By default, the packet to 20.1.10.2 will be routed to server1, gets tunneled to the pod on server0, and finally returns to router directly using DSR.
You can observe this by running tcpdump at right interfaces:

On server1, you can see the packet with LoadBalancerIP as dst IP come in.
```sh
$ sudo nsenter -t $(docker inspect -f '{{.State.Pid}}' clab-bgpv2-bird-coil-server1) -n tcpdump tcp and host 20.1.10.2 -i net0
tcpdump: verbose output suppressed, use -v[v]... for full protocol decode
listening on net0, link-type EN10MB (Ethernet), snapshot length 262144 bytes
13:03:33.068103 IP 10.0.1.1.50001 > 20.1.10.2.http: Flags [S], seq 570850030, win 56760, options [mss 9460,sackOK,TS val 2396702909 ecr 0,nop,wscale 7], length 0
13:03:33.068239 IP 10.0.1.1.50001 > 20.1.10.2.http: Flags [.], ack 2860026207, win 444, options [nop,nop,TS val 2396702909 ecr 2789094429], length 0
13:03:33.070735 IP 10.0.1.1.50001 > 20.1.10.2.http: Flags [P.], seq 0:73, ack 1, win 444, options [nop,nop,TS val 2396702911 ecr 2789094429], length 73: HTTP: GET / HTTP/1.1
13:03:33.070892 IP 10.0.1.1.50001 > 20.1.10.2.http: Flags [.], ack 194, win 443, options [nop,nop,TS val 2396702911 ecr 2789094431], length 0
13:03:33.073949 IP 10.0.1.1.50001 > 20.1.10.2.http: Flags [F.], seq 73, ack 194, win 443, options [nop,nop,TS val 2396702914 ecr 2789094431], length 0
13:03:33.074083 IP 10.0.1.1.50001 > 20.1.10.2.http: Flags [.], ack 195, win 443, options [nop,nop,TS val 2396702915 ecr 2789094434], length 0
```

On server0, you can see the packet with LoadBalancerIP as src IP go out.
```sh
$ sudo nsenter -t $(docker inspect -f '{{.State.Pid}}' clab-bgpv2-bird-coil-server0) -n tcpdump tcp and host 20.1.10.2 -i net0
tcpdump: verbose output suppressed, use -v[v]... for full protocol decode
listening on net0, link-type EN10MB (Ethernet), snapshot length 262144 bytes
13:03:33.068230 IP 20.1.10.2.http > 10.0.1.1.50001: Flags [S.], seq 2860026206, ack 570850031, win 56688, options [mss 9460,sackOK,TS val 2789094429 ecr 2396702909,nop,wscale 7], length 0
13:03:33.070780 IP 20.1.10.2.http > 10.0.1.1.50001: Flags [.], ack 74, win 443, options [nop,nop,TS val 2789094431 ecr 2396702911], length 0
13:03:33.070886 IP 20.1.10.2.http > 10.0.1.1.50001: Flags [P.], seq 1:194, ack 74, win 443, options [nop,nop,TS val 2789094431 ecr 2396702911], length 193: HTTP: HTTP/1.1 200 OK
13:03:33.074074 IP 20.1.10.2.http > 10.0.1.1.50001: Flags [F.], seq 194, ack 75, win 443, options [nop,nop,TS val 2789094434 ecr 2396702914], length 0
```


By default, ECMP on router is computed only by L3 hash (src IP and dst IP). To effectively simulate ECMP, use L4 hash (5-tuples) by:
```sh
sudo nsenter -t $(docker inspect -f '{{.State.Pid}}' clab-bgpv2-bird-coil-router0) -n sysctl -w net.ipv4.fib_multipath_hash_policy=1
# Then curl the LoadBalancer IP with different local ports
docker exec clab-bgpv2-bird-coil-router0 curl --local-port 20000 -sSL 20.1.10.2
docker exec clab-bgpv2-bird-coil-router0 curl --local-port 60000 -sSL 20.1.10.2

# Check the pod logs to see that the source IP changes based on the path selected.
stern -n tenant-green deploy/ubuntu-echo
+ ubuntu-echo-854bb9b8d6-drpdg › ubuntu
ubuntu-echo-854bb9b8d6-drpdg ubuntu 2025/07/04 00:53:12 [INFO] server is listening on :8080
ubuntu-echo-854bb9b8d6-drpdg ubuntu 2025/07/04 01:44:58 20.1.10.2 10.0.1.1:20000 "GET / HTTP/1.1" 200 31 "curl/7.81.0" 3.667µs
ubuntu-echo-854bb9b8d6-drpdg ubuntu 2025/07/04 01:45:04 20.1.10.2 10.0.2.1:60000 "GET / HTTP/1.1" 200 31 "curl/7.81.0" 4.459µs
```

### Cleanup
```sh
# Destroy
make destroy

# Reload (Destroy then re-deploy)
make reload

```

