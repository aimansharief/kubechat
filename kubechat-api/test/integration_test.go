//go:build integration

package test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"
)

const baseURL = "http://localhost:8080/api/v1"

func setupKindCluster(t *testing.T) {
	t.Helper()
	cmd := exec.Command("kind", "create", "cluster", "--name", "kubechat-test")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to create kind cluster: %v", err)
	}
	t.Cleanup(func() {
		exec.Command("kind", "delete", "cluster", "--name", "kubechat-test").Run()
	})
}

func postCommand(t *testing.T, command string) *http.Response {
	body := map[string]interface{}{"command": command}
	b, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/execute", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /execute error: %v", err)
	}
	return resp
}

func decodeBody(t *testing.T, resp *http.Response, out interface{}) {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
}

func TestIntegration_CRUDAndSecurity(t *testing.T) {
	setupKindCluster(t)
	t.Run("BlockedCommands", func(t *testing.T) {
		resp := postCommand(t, "kubectl delete pod foo")
		if resp.StatusCode != 403 {
			t.Errorf("Expected 403 for delete, got %d", resp.StatusCode)
		}
	})
	t.Run("AllowedGetPods", func(t *testing.T) {
		resp := postCommand(t, "kubectl get pods -n default")
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200 for get pods, got %d", resp.StatusCode)
		}
		var body map[string]interface{}
		decodeBody(t, resp, &body)
		if _, ok := body["output"]; !ok {
			t.Errorf("Expected output field in response")
		}
	})
}

func TestIntegration_ClusterHealth(t *testing.T) {
	setupKindCluster(t)
	resp, err := http.Get(baseURL + "/cluster-health")
	if err != nil {
		t.Fatalf("Error calling cluster-health: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	decodeBody(t, resp, &body)
	if _, ok := body["nodes"]; !ok {
		t.Errorf("Expected nodes in health response")
	}
}

func TestIntegration_CoverageReport(t *testing.T) {
	// This is a placeholder to remind to run:
	// go test -coverprofile=coverage.out ./...
}
