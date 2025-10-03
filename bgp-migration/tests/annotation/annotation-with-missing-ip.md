# Test if lbipam.cilium.io/ips annotation is required

Mostly followed the basic steps defiend in ./annotation.md.
One change is that after `make deploy` and `make remove-annotation`, create additional LB Services and delete one.
```sh
kubectl apply -f test_new_lbs.yaml
kubectl delete svc default-lb-two
```

## Result:
TL;DR: Annotation may not be needed.
### Service Watch
No LBIP re-assignments.
```sh
s
NAMESPACE   NAME               TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
bastion     bastion-lb         LoadBalancer   10.2.221.161   20.1.20.0     80:32648/TCP   104s
bastion     bastion-lb-two     LoadBalancer   10.2.214.96    20.1.20.1     80:32467/TCP   104s
default     bastion-lb-three   LoadBalancer   10.2.17.208    20.1.20.2     80:31003/TCP   73s
default     default-lb         LoadBalancer   10.2.68.43     20.1.10.0     80:32488/TCP   104s
default     default-lb-three   LoadBalancer   10.2.7.147     20.1.10.2     80:31653/TCP   73s

apply patch & bgp cp

apply lbippool

apply bgp cp

restart agents 1

restart agents 2

restart agents 3

```
### BGP advertisement watch
Omitted since it was same as the test with and without annotation, except the difference in the set of LBIPs advertised.
