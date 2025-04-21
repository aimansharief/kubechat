package api

import (
	"net/http"
	"kubechat-api/config"
	"kubechat-api/kube"
	"kubechat-api/auditlog"
	"fmt"
	"strings"
	"time"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"io"
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, cfg config.Config, logger *zap.Logger, kubeClient *kube.KubeClient) {

	v1 := r.Group("/api/v1")
	{
		v1.POST("/parse", parseHandler)
		v1.POST("/execute", KubectlValidator(cfg, kubeClient, logger), executeHandler)
		v1.POST("/dry-run", KubectlValidator(cfg, kubeClient, logger), dryRunHandler)
		v1.GET("/context", contextHandler)
		v1.GET("/insights", insightsHandler)
		v1.GET("/suggestions", suggestionsHandler)
		v1.GET("/metrics", metricsHandler)
		v1.GET("/health", healthHandler)
		v1.GET("/cluster-health", ClusterHealthHandler(kubeClient, "dev-cluster"))
	}
}

func parseHandler(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)

	// Simple NLP mapping (stub): map common phrases to kubectl commands
	var req struct {
		Query string `json:"query" binding:"required,min=3,max=500"`
	}
	requestID := c.GetString("requestID")

	if c.Request.ContentLength > 1024*1024 {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":   "Payload too large",
			"code":    "ERR_PAYLOAD_TOO_LARGE",
			"details": "Request exceeds 1MB limit",
		})
		return
	}

	// Use cached body from context and log
	if v, exists := c.Get("rawBody"); exists {
		if bodyBytes, ok := v.([]byte); ok {
			logger.Info("[DEBUG] Raw request body in parseHandler", zap.ByteString("body", bodyBytes))
			logger.Info("[DEBUG] Request headers in parseHandler", zap.Any("headers", c.Request.Header))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "ERR_INVALID_INPUT",
			"details": err.Error(),
			"request_id": requestID,
			"locale": "en-US",
		})
		return
	}
	if len(req.Query) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Query too long",
			"code":    "ERR_QUERY_TOO_LONG",
			"details": "Query exceeds 500 characters",
			"request_id": requestID,
			"locale": "en-US",
		})
		return
	}
	if hasInjection(req.Query) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Potential command injection detected",
			"code":    "ERR_COMMAND_INJECTION",
			"details": req.Query,
			"request_id": requestID,
			"locale": "en-US",
		})
		return
	}
	// Simple NLP mapping
	mapping := map[string]string{
		"scale frontend to 3 replicas": "kubectl scale deployment/frontend --replicas=3",
		"show pods": "kubectl get pods",
		"list namespaces": "kubectl get namespaces",
		"describe pod frontend": "kubectl describe pod frontend",
	}
	cmd, found := mapping[strings.ToLower(strings.TrimSpace(req.Query))]
	if !found {
		cmd = "kubectl get pods" // fallback
	}
	c.JSON(http.StatusOK, gin.H{
		"command":   cmd,
		"confidence": 0.90,
		"alternatives": []string{},
		"requires_confirmation": true,
		"request_id": requestID,
	})
}

func executeHandler(c *gin.Context) {
	// --- Setup and parse request ---
	requestID := c.GetString("requestID")
	logger := c.MustGet("logger").(*zap.Logger)
	var req struct {
		Command string `json:"command" binding:"required,min=3,max=500"`
		DryRun  bool   `json:"dry_run"`
	}
	// Use cached body from context
	if v, exists := c.Get("rawBody"); exists {
		if bodyBytes, ok := v.([]byte); ok {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("[DEBUG] ShouldBindJSON error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "ERR_INVALID_INPUT",
			"details": err.Error(),
			"request_id": requestID,
			"locale": "en-US",
		})
		return
	}
	if len(req.Command) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Command too long",
			"code":    "ERR_COMMAND_TOO_LONG",
			"details": "Command exceeds 500 characters",
			"request_id": requestID,
			"locale": "en-US",
		})
		return
	}
	// --- Use kubeClient from Gin context ---
	kubeClientIface, exists := c.Get("kubeClient")
	if !exists || kubeClientIface == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kubernetes client not available"})
		return
	}
	kubeClient, ok := kubeClientIface.(*kube.KubeClient)
	if !ok || kubeClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Kubernetes client"})
		return
	}
	// Get cluster name from config (first cluster or unknown)
	cfg := config.Load("config/default.yaml")
	clusterName := "unknown"
	if len(cfg.Clusters) > 0 {
		clusterName = cfg.Clusters[0].Name
	}

	// --- Parse command ---
	parts := strings.Fields(req.Command)
	if len(parts) < 3 || parts[0] != "kubectl" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid kubectl command syntax"})
		return
	}
	verb := parts[1]
	resource := parts[2]
	var namespace string = "default"
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "-n" && i+1 < len(parts) {
			namespace = parts[i+1]
		}
	}
	output := ""
	success := true
	var opErr error
	var dryRun bool = req.DryRun
	switch verb {
	case "get":
		if resource == "pods" {
			pods, err := kubeClient.Clientset.CoreV1().Pods(namespace).List(c, metav1.ListOptions{})
			if err != nil {
				success = false
				switch {
				case strings.Contains(err.Error(), "forbidden"):
					output = "Error from server (Forbidden): RBAC: access denied"
				case strings.Contains(err.Error(), "not found"):
					output = "Error from server (NotFound): pods not found"
				case strings.Contains(err.Error(), "no such namespace"):
					output = "Error from server (NotFound): namespace not found"
				default:
					output = "Error: " + err.Error()
				}
			} else {
				output = "NAME\tREADY\tSTATUS\tRESTARTS\tAGE\n"
				for _, p := range pods.Items {
					output += p.Name + "\t"
					readyVal := 0
					if len(p.Status.ContainerStatuses) > 0 && p.Status.ContainerStatuses[0].Ready {
						readyVal = 1
					}
					output += fmt.Sprintf("%d/%d\t", readyVal, len(p.Status.ContainerStatuses))
					output += string(p.Status.Phase) + "\t"
					if len(p.Status.ContainerStatuses) > 0 {
						output += fmt.Sprintf("%d\t", p.Status.ContainerStatuses[0].RestartCount)
					} else {
						output += "0\t"
					}
					output += fmt.Sprintf("%s\n", p.CreationTimestamp.Time.Format("2m"))
				}
			}
		}
		// Add more resources as needed
	case "delete":
		if dryRun {
			output = "pod deleted (dry-run)"
		} else {
			opErr = fmt.Errorf("delete not implemented in mock")
		}
	case "scale":
		if dryRun {
			output = "deployment scaled (dry-run)"
		} else {
			// Parse deployment name and replicas
			var deploymentName string
			replicas := int32(1)
			for i := 3; i < len(parts); i++ {
				if strings.HasPrefix(parts[i], "--replicas=") {
					replicaStr := strings.TrimPrefix(parts[i], "--replicas=")
					if n, err := fmt.Sscanf(replicaStr, "%d", &replicas); n == 1 && err == nil {
						// parsed successfully
					} else {
						opErr = fmt.Errorf("invalid replicas argument: %s", replicaStr)
						break
					}
				}
			}
			// deploymentName is after 'deployment/'
			if strings.HasPrefix(resource, "deployment/") {
				deploymentName = strings.TrimPrefix(resource, "deployment/")
			} else if len(parts) > 3 && strings.HasPrefix(parts[3], "deployment/") {
				deploymentName = strings.TrimPrefix(parts[3], "deployment/")
			}
			if deploymentName == "" {
				opErr = fmt.Errorf("could not parse deployment name")
			} else if opErr == nil {
				err := kubeClient.ScaleDeployment(c, namespace, deploymentName, replicas)
				if err != nil {
					opErr = err
				} else {
					output = fmt.Sprintf("deployment/%s scaled to %d replicas", deploymentName, replicas)
				}
			}
		}
	default:
		opErr = fmt.Errorf("unsupported verb: %s", verb)
	}
	if opErr != nil {
		success = false
		output = opErr.Error()
	}
	// --- Audit log ---
	userID := c.GetString("userID") // must be set by JWT middleware
	auditlog.AuditLog(logger, auditlog.AuditEntry{
		Timestamp: time.Now(),
		UserID:    userID,
		Cluster:   clusterName,
		Command:   req.Command,
		Success:   success,
		Details:   output,
	})
	// --- Response ---
	c.JSON(http.StatusOK, gin.H{
		"output": output,
		"cluster": clusterName,
		"executed_at": time.Now().Format(time.RFC3339),
		"dry_run": dryRun,
	})
	return
}

func metricsHandler(c *gin.Context) {
	// Return basic metrics about the cluster (stub, but structure for real impl)
	kubeClientIface, exists := c.Get("kubeClient")
	var podCount int
	if exists && kubeClientIface != nil {
		kubeClient, ok := kubeClientIface.(*kube.KubeClient)
		if ok && kubeClient != nil {
			podList, err := kubeClient.Clientset.CoreV1().Pods("").List(c, metav1.ListOptions{})
			if err == nil {
				podCount = len(podList.Items)
			}
		}
	}
	requestID := c.GetString("requestID")
	c.JSON(http.StatusOK, gin.H{
		"pods": podCount,
		"cpu":   "N/A",
		"memory": "N/A",
		"health": "OK",
		"request_id": requestID,
	})
}

func insightsHandler(c *gin.Context) {
	// Return cluster insights (stub, but structure for real impl)
	kubeClientIface, exists := c.Get("kubeClient")
	var podWarnings []gin.H
	if exists && kubeClientIface != nil {
		kubeClient, ok := kubeClientIface.(*kube.KubeClient)
		if ok && kubeClient != nil {
			podList, err := kubeClient.Clientset.CoreV1().Pods("").List(c, metav1.ListOptions{})
			if err == nil {
				for _, pod := range podList.Items {
					if pod.Status.Phase == "CrashLoopBackOff" {
						podWarnings = append(podWarnings, gin.H{
							"type": "CrashLoopBackOff",
							"message": fmt.Sprintf("Pod %s is restarting frequently", pod.Name),
							"severity": "high",
							"suggestion": fmt.Sprintf("Check logs with: kubectl logs %s", pod.Name),
						})
					}
				}
			}
		}
	}
	requestID := c.GetString("requestID")
	c.JSON(http.StatusOK, gin.H{
		"insights": podWarnings,
		"request_id": requestID,
	})
}

func healthHandler(c *gin.Context) {
	requestID := c.GetString("requestID")
	c.JSON(http.StatusOK, gin.H{"status": "ok", "request_id": requestID})
}
