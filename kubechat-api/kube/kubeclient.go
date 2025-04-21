package kube

import (
	"context"
	"fmt"
	"encoding/json"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

// KubeClient wraps a Kubernetes clientset and config
// Supports auto-refresh and health checks
//
type KubeClient struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	kubeconfigPath string
	inCluster bool
}

// NewKubeClient loads config from file or in-cluster
func NewKubeClient(kubeconfigPath string, tryInCluster bool) (*KubeClient, error) {
	var cfg *rest.Config
	var err error
	if kubeconfigPath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else if tryInCluster {
		cfg, err = rest.InClusterConfig()
	} else {
		return nil, fmt.Errorf("no kubeconfig path provided and not in-cluster")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	return &KubeClient{Clientset: clientset, Config: cfg, kubeconfigPath: kubeconfigPath, inCluster: tryInCluster}, nil
}

// HealthCheck checks cluster connectivity (e.g., can list namespaces)
func (kc *KubeClient) HealthCheck(ctx context.Context) error {
	_, err := kc.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	return err
}

// ScaleDeployment scales a deployment to the desired number of replicas
func (kc *KubeClient) ScaleDeployment(ctx context.Context, namespace, deploymentName string, replicas int32) error {
	// Prepare patch
	replicaPatch := map[string]interface{}{
		"spec": map[string]interface{}{
			"replicas": replicas,
		},
	}
	patchBytes, err := json.Marshal(replicaPatch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}
	_, err = kc.Clientset.AppsV1().Deployments(namespace).Patch(ctx, deploymentName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}
	return nil
}

// Refresh reloads credentials (for expired tokens, etc)
func (kc *KubeClient) Refresh() error {
	var cfg *rest.Config
	var err error
	if kc.kubeconfigPath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kc.kubeconfigPath)
	} else if kc.inCluster {
		cfg, err = rest.InClusterConfig()
	} else {
		return fmt.Errorf("no kubeconfig path or in-cluster config")
	}
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}
	kc.Config = cfg
	kc.Clientset = clientset
	return nil
}
