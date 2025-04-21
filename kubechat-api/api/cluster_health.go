package api

import (
	"sync"
	"time"
	"net/http"
	"kubechat-api/kube"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterHealthCache struct {
	mu sync.Mutex
	lastResult map[string]interface{}
	timestamp time.Time
}

var healthCache = &ClusterHealthCache{}

func ClusterHealthHandler(kubeClient *kube.KubeClient, clusterName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		healthCache.mu.Lock()
		if healthCache.lastResult != nil && time.Since(healthCache.timestamp) < 30*time.Second {
			resp := healthCache.lastResult
			healthCache.mu.Unlock()
			c.JSON(http.StatusOK, resp)
			return
		}
		healthCache.mu.Unlock()

		status := map[string]interface{}{
			"cluster": clusterName,
			"healthy": true,
			"nodes": map[string]int{"total": 0, "ready": 0},
			"system_components": map[string]string{"api_server": "unknown", "scheduler": "unknown"},
		}
		// API server reachability
		err := kubeClient.HealthCheck(c)
		if err != nil {
			status["healthy"] = false
			status["system_components"].(map[string]string)["api_server"] = "unreachable"
		} else {
			status["system_components"].(map[string]string)["api_server"] = "ok"
		}
		// Node readiness
		nodes, err := kubeClient.Clientset.CoreV1().Nodes().List(c, metav1.ListOptions{})
		totalNodes, readyNodes := 0, 0
		if err == nil {
			totalNodes = len(nodes.Items)
			for _, node := range nodes.Items {
				for _, cond := range node.Status.Conditions {
					if cond.Type == "Ready" && cond.Status == "True" {
						readyNodes++
						break
					}
				}
			}
			status["nodes"] = map[string]int{"total": totalNodes, "ready": readyNodes}
			if readyNodes < totalNodes {
				status["healthy"] = false
			}
		}
		// Pod capacity
		pods, err := kubeClient.Clientset.CoreV1().Pods("").List(c, metav1.ListOptions{})
		if err == nil {
			status["pods_total"] = len(pods.Items)
		}
		// Critical system pods status
		schedulerStatus := "unknown"
		for _, pod := range pods.Items {
			if pod.Namespace == "kube-system" && pod.Name != "" {
				if pod.Labels["component"] == "kube-scheduler" {
					if pod.Status.Phase == "Running" {
						schedulerStatus = "ok"
					} else {
						schedulerStatus = string(pod.Status.Phase)
					}
				}
			}
		}
		status["system_components"].(map[string]string)["scheduler"] = schedulerStatus
		// Cache result
		healthCache.mu.Lock()
		healthCache.lastResult = status
		healthCache.timestamp = time.Now()
		healthCache.mu.Unlock()
		c.JSON(http.StatusOK, status)
	}
}
