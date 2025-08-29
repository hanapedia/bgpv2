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

const annotationKey = "lbipam.cilium.io/ips"

func main() {
	var kubeconfig string
	var remove bool

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file (optional; if empty, try in-cluster)")
	flag.BoolVar(&remove, "remove", false, "remove the lbipam.cilium.io/ips annotation instead of setting it")
	flag.Parse()

	// Build config: try in-cluster first, then fall back to kubeconfig
	var cfg *rest.Config
	var err error
	if cfg, err = rest.InClusterConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "🟢 using in-cluster configuration")
	} else {
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

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ failed to create clientset: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	services, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ error listing services: %v\n", err)
		os.Exit(1)
	}

	for _, svc := range services.Items {
		// operate only on LoadBalancer services
		if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
			continue
		}

		if remove {
			// skip if annotation isn't present
			if svc.Annotations == nil {
				continue
			}
			if _, ok := svc.Annotations[annotationKey]; !ok {
				continue
			}

			// JSON Merge Patch with null deletes the key
			patch := map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						annotationKey: nil,
					},
				},
			}
			patchBytes, err := json.Marshal(patch)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to marshal remove-patch for %s/%s: %v\n", svc.Namespace, svc.Name, err)
				continue
			}
			_, err = clientset.CoreV1().Services(svc.Namespace).Patch(ctx, svc.Name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to remove annotation on %s/%s: %v\n", svc.Namespace, svc.Name, err)
				continue
			}
			fmt.Printf("🗑 removed %s from %s/%s\n", annotationKey, svc.Namespace, svc.Name)
			continue
		}

		// collect external IPs (IP or hostname)
		var ips []string
		for _, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				ips = append(ips, ing.IP)
			} else if ing.Hostname != "" {
				ips = append(ips, ing.Hostname)
			}
		}
		if len(ips) == 0 {
			// nothing to write
			continue
		}

		patch := map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": map[string]string{
					annotationKey: strings.Join(ips, ","),
				},
			},
		}
		patchBytes, err := json.Marshal(patch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal patch for %s/%s: %v\n", svc.Namespace, svc.Name, err)
			continue
		}

		_, err = clientset.CoreV1().Services(svc.Namespace).Patch(ctx, svc.Name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to patch %s/%s: %v\n", svc.Namespace, svc.Name, err)
			continue
		}
		fmt.Printf("✅ patched %s/%s → %s=%q\n", svc.Namespace, svc.Name, annotationKey, strings.Join(ips, ","))
	}
}
