# Test if lbipam.cilium.io/ips annotation is required

Mostly followed the basic steps defiend in ../procedures.md.
One change is that after `make deploy`, ran `make remove-annotation` to strip away the annotation.

## Result:
TL;DR: Annotation may not be needed.
**TODO:**
- [x] try with missing LB IP in the middle of the pool since all the IPs were sequential
    - see ./annotation-with-missing-ip.md
- [ ] confirm at the source code level

### Service Watch
No LBIP re-assignments.
```sh
$ k get svc -A --field-selector="spec.type=LoadBalancer" -w
NAMESPACE   NAME             TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
bastion     bastion-lb       LoadBalancer   10.2.77.173    20.1.20.0     80:30171/TCP   56s
bastion     bastion-lb-two   LoadBalancer   10.2.199.35    20.1.20.1     80:31470/TCP   56s
default     default-lb       LoadBalancer   10.2.224.101   20.1.10.0     80:32712/TCP   56s
default     default-lb-two   LoadBalancer   10.2.91.205    20.1.10.1     80:32238/TCP   56s

# install bgp-cp

# apply lbipam

# apply bgp cp

# restart agents
# 1

# 2

# 3
```
### BGP advertisement watch
Same as the test with annotation.
```sh
$ docker logs clab-bgp-migration-router0 -f | grep -e '20.1.10.' -e '20.1.20.'
bird: importing route: 20.1.20.1/32, 10.0.3.2, worker2, 10.0.3.2
bird: worker2.ipv4 > added [best] 20.1.20.1/32 unicast
bird: importing route: 20.1.20.0/32, 10.0.3.2, worker2, 10.0.3.2
bird: worker2.ipv4 > added [best] 20.1.20.0/32 unicast
bird: importing route: 20.1.20.1/32, 10.0.2.2, worker, 10.0.2.2
bird: worker.ipv4 > added [best] 20.1.20.1/32 unicast
bird: importing route: 20.1.20.0/32, 10.0.2.2, worker, 10.0.2.2
bird: worker.ipv4 > added [best] 20.1.20.0/32 unicast
bird: importing route: 20.1.20.0/32, 10.0.1.2, controlplane, 10.0.1.2
bird: controlplane.ipv4 > added [best] 20.1.20.0/32 unicast
bird: importing route: 20.1.20.1/32, 10.0.1.2, controlplane, 10.0.1.2
bird: controlplane.ipv4 > added [best] 20.1.20.1/32 unicast
bird: importing route: 20.1.10.0/32, 10.0.2.2, worker, 10.0.2.2
bird: worker.ipv4 > added [best] 20.1.10.0/32 unicast
bird: importing route: 20.1.10.0/32, 10.0.1.2, controlplane, 10.0.1.2
bird: controlplane.ipv4 > added [best] 20.1.10.0/32 unicast
bird: importing route: 20.1.10.0/32, 10.0.3.2, worker2, 10.0.3.2
bird: worker2.ipv4 > added 20.1.10.0/32 unicast
bird: importing route: 20.1.10.1/32, 10.0.1.2, controlplane, 10.0.1.2
bird: controlplane.ipv4 > added [best] 20.1.10.1/32 unicast
bird: importing route: 20.1.10.1/32, 10.0.2.2, worker, 10.0.2.2
bird: worker.ipv4 > added 20.1.10.1/32 unicast
bird: importing route: 20.1.10.1/32, 10.0.3.2, worker2, 10.0.3.2
bird: worker2.ipv4 > added 20.1.10.1/32 unicast
bird: worker.ipv4 > removed 20.1.10.1/32 unicast
bird: worker.ipv4 > removed 20.1.20.1/32 unicast
bird: worker.ipv4 > removed 20.1.10.0/32 unicast
bird: worker.ipv4 > removed 20.1.20.0/32 unicast
bird: importing route: 20.1.10.1/32, 10.0.2.2, worker, 10.0.2.2
bird: worker.ipv4 > added 20.1.10.1/32 unicast
bird: importing route: 20.1.20.0/32, 10.0.2.2, worker, 10.0.2.2
bird: worker.ipv4 > added 20.1.20.0/32 unicast
bird: importing route: 20.1.20.1/32, 10.0.2.2, worker, 10.0.2.2
bird: worker.ipv4 > added 20.1.20.1/32 unicast
bird: importing route: 20.1.10.0/32, 10.0.2.2, worker, 10.0.2.2
bird: worker.ipv4 > added 20.1.10.0/32 unicast
bird: controlplane.ipv4 > removed [replaced] 20.1.10.1/32 unicast
bird: controlplane.ipv4 > removed [replaced] 20.1.20.1/32 unicast
bird: controlplane.ipv4 > removed [replaced] 20.1.10.0/32 unicast
bird: controlplane.ipv4 > removed [replaced] 20.1.20.0/32 unicast
bird: importing route: 20.1.20.0/32, 10.0.1.2, controlplane, 10.0.1.2
bird: controlplane.ipv4 > added [best] 20.1.20.0/32 unicast
bird: importing route: 20.1.20.1/32, 10.0.1.2, controlplane, 10.0.1.2
bird: controlplane.ipv4 > added [best] 20.1.20.1/32 unicast
bird: importing route: 20.1.10.0/32, 10.0.1.2, controlplane, 10.0.1.2
bird: controlplane.ipv4 > added [best] 20.1.10.0/32 unicast
bird: importing route: 20.1.10.1/32, 10.0.1.2, controlplane, 10.0.1.2
bird: controlplane.ipv4 > added [best] 20.1.10.1/32 unicast
bird: worker2.ipv4 > removed 20.1.10.1/32 unicast
bird: worker2.ipv4 > removed 20.1.20.1/32 unicast
bird: worker2.ipv4 > removed 20.1.10.0/32 unicast
bird: worker2.ipv4 > removed 20.1.20.0/32 unicast
bird: importing route: 20.1.20.1/32, 10.0.3.2, worker2, 10.0.3.2
bird: worker2.ipv4 > added 20.1.20.1/32 unicast
bird: importing route: 20.1.10.0/32, 10.0.3.2, worker2, 10.0.3.2
bird: worker2.ipv4 > added 20.1.10.0/32 unicast
bird: importing route: 20.1.10.1/32, 10.0.3.2, worker2, 10.0.3.2
bird: worker2.ipv4 > added 20.1.10.1/32 unicast
bird: importing route: 20.1.20.0/32, 10.0.3.2, worker2, 10.0.3.2
bird: worker2.ipv4 > added 20.1.20.0/32 unicast
```
