package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"kubechat-api/config"
	"kubechat-api/kube"
	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	logger, _ := zap.NewDevelopment()
	cfg := config.Load("../config/default.yaml")
	var kubeClient *kube.KubeClient = nil // Use nil for unit tests
	r := gin.Default()
	RegisterRoutes(r, cfg, logger, kubeClient)
	return r
}

func TestParseEndpoint(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	body := map[string]interface{}{ "query": "Scale frontend to 3 replicas" }
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/parse", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["command"]; !ok {
		t.Errorf("Expected command in response")
	}
}

func TestDryRunEndpoint(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	body := map[string]interface{}{ "command": "kubectl get pods" }
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/dry-run", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["result"]; !ok {
		t.Errorf("Expected result in response")
	}
}

func TestContextEndpoint(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/context", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["namespaces"]; !ok {
		t.Errorf("Expected namespaces in response")
	}
}

func TestSuggestionsEndpoint(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/suggestions?resource=frontend", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["suggestions"]; !ok {
		t.Errorf("Expected suggestions in response")
	}
}

func TestHealthEndpoint(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["status"]; !ok {
		t.Errorf("Expected status in response")
	}
}
