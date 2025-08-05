// main.go
package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
)

func main() {
    // 1) Parse a --kubeconfig flag (optional)
    var kubeconfig string
    flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file (optional; if empty, try in-cluster)")
    flag.Parse()

    // 2) Build a REST config: try in-cluster first, then fall back to kubeconfig
    var cfg *rest.Config
    var err error

    if cfg, err = rest.InClusterConfig(); err == nil {
        fmt.Fprintf(os.Stderr, "🟢 using in-cluster configuration\n")
    } else {
        // no in-cluster; fall back
        if kubeconfig == "" {
            if home := homedir.HomeDir(); home != "" {
                kubeconfig = filepath.Join(home, ".kube", "config")
            }
        }
        fmt.Fprintf(os.Stderr, "⚪️ falling back to kubeconfig=%q\n", kubeconfig)
        cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            fmt.Fprintf(os.Stderr, "❌ failed to build config from kubeconfig %q: %v\n", kubeconfig, err)
            os.Exit(1)
        }
    }

    // 3) Create the clientset
    clientset, err := kubernetes.NewForConfig(cfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ failed to create clientset: %v\n", err)
        os.Exit(1)
    }

    ctx := context.Background()

    // 4) List ALL Services in ALL namespaces
    services, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ error listing services: %v\n", err)
        os.Exit(1)
    }

    // 5) Iterate and patch
    for _, svc := range services.Items {
        if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
            continue
        }

        // collect external IPs or hostnames
        var ips []string
        for _, ing := range svc.Status.LoadBalancer.Ingress {
            if ing.IP != "" {
                ips = append(ips, ing.IP)
            } else if ing.Hostname != "" {
                ips = append(ips, ing.Hostname)
            }
        }
        if len(ips) == 0 {
            continue
        }

        // build the patch
        patch := map[string]interface{}{
            "metadata": map[string]interface{}{
                "annotations": map[string]string{
                    "lbipam.cilium.io/ips": strings.Join(ips, ","),
                },
            },
        }
        data, err := json.Marshal(patch)
        if err != nil {
            fmt.Fprintf(os.Stderr, "❌ marshal patch for %s/%s: %v\n", svc.Namespace, svc.Name, err)
            continue
        }

        // apply it
        if _, err := clientset.CoreV1().
            Services(svc.Namespace).
            Patch(ctx, svc.Name, types.MergePatchType, data, metav1.PatchOptions{}); err != nil {
            fmt.Fprintf(os.Stderr, "❌ patch %s/%s: %v\n", svc.Namespace, svc.Name, err)
            continue
        }

        fmt.Printf("✅ patched %s/%s → ips=%q\n", svc.Namespace, svc.Name, strings.Join(ips, ","))
    }
}
