package kube

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewKubeClient_ValidKubeconfig(t *testing.T) {
	path := os.Getenv("KUBECONFIG")
	if path == "" {
		t.Skip("KUBECONFIG not set")
	}
	_, err := NewKubeClient(path, false)
	if err != nil {
		t.Fatalf("Expected valid kubeconfig, got error: %v", err)
	}
}

func TestNewKubeClient_InvalidKubeconfig(t *testing.T) {
	_, err := NewKubeClient("/invalid/path/to/kubeconfig", false)
	if err == nil {
		t.Fatal("Expected error for invalid kubeconfig, got nil")
	}
}

func TestNewKubeClient_InClusterFallback(t *testing.T) {
	// This will likely fail unless running in cluster, so expect error
	_, err := NewKubeClient("", true)
	if err == nil {
		t.Fatal("Expected error for missing in-cluster config, got nil")
	}
}

func TestHealthCheck_NetworkFailure(t *testing.T) {
	// Use an unreachable API server
	os.Setenv("KUBECONFIG", "testdata/unreachable.kubeconfig")
	client, err := NewKubeClient("testdata/unreachable.kubeconfig", false)
	if err != nil {
		t.Skip("Cannot create client for network failure test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.HealthCheck(ctx)
	if err == nil {
		t.Fatal("Expected network failure error, got nil")
	}
}

func TestHealthCheck_RBACDenied(t *testing.T) {
	// This test assumes a kubeconfig/user with no list namespace permission
	path := os.Getenv("KUBECONFIG_DENIED")
	if path == "" {
		t.Skip("KUBECONFIG_DENIED not set")
	}
	client, err := NewKubeClient(path, false)
	if err != nil {
		t.Fatalf("Failed to load client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.HealthCheck(ctx)
	if err == nil {
		t.Fatal("Expected RBAC denial error, got nil")
	}
}
