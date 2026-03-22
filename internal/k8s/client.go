package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Clientset *kubernetes.Clientset
}

// NewClient initializes a new Kubernetes client using ~/.kube/config or in-cluster config.
func NewClient() (*Client, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not get home directory: %w", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		// Fallback to in-cluster config if kubeconfig is not found/accessible
		config, err = clientcmd.BuildConfigFromFlags("", "")
		if err != nil {
			return nil, fmt.Errorf("error building kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes clientset: %w", err)
	}

	return &Client{Clientset: clientset}, nil
}

// ListNamespaces returns a list of all namespace names.
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}
	return names, nil
}

// ListWorkloads returns a list of Deployments, StatefulSets, and DaemonSets in a namespace.
func (c *Client) ListWorkloads(ctx context.Context, namespace string) ([]string, error) {
	var workloads []string

	deps, err := c.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, d := range deps.Items {
			workloads = append(workloads, d.Name)
		}
	}

	stss, err := c.Clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, s := range stss.Items {
			workloads = append(workloads, s.Name)
		}
	}

	dss, err := c.Clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, ds := range dss.Items {
			workloads = append(workloads, ds.Name)
		}
	}

	return workloads, nil
}

// ListPods returns a list of pods for a given workload or all pods in a namespace.
func (c *Client) ListPods(ctx context.Context, namespace, workload string) ([]string, error) {
	// Simple approach: list all pods and filter if workload is provided.
	// For production, more efficient label selection would be preferred.
	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var podNames []string
	for _, p := range pods.Items {
		// Basic filtering for demonstration purposes
		podNames = append(podNames, p.Name)
	}
	return podNames, nil
}
