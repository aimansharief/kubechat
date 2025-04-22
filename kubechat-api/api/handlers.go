package api

import (
	"net/http"
	"kubechat-api/config"
	"kubechat-api/kube"
	"kubechat-api/auditlog"
	"fmt"
	"strings"
	"time"
	"io"
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

func RegisterRoutes(r *gin.Engine, cfg config.Config, logger *zap.Logger, kubeClient *kube.KubeClient) {

	v1 := r.Group("/api/v1")
	{
				v1.POST("/execute", KubectlValidator(cfg, kubeClient, logger), executeHandler)
		v1.POST("/dry-run", KubectlValidator(cfg, kubeClient, logger), dryRunHandler)
		v1.POST("/llm-parse", llmParseHandler)
		v1.GET("/context", contextHandler)
		v1.GET("/insights", insightsHandler)
				v1.GET("/metrics", metricsHandler)
		v1.GET("/health", healthHandler)
		v1.GET("/cluster-health", ClusterHealthHandler(kubeClient, "dev-cluster"))
	}
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
		logger.Error("Kubernetes client not available in context. This usually means kubeconfig is missing or invalid.")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Kubernetes client not available. Check kubeconfig setup on server.",
			"hint": "Ensure kubeconfig.yaml exists and is valid. See README for details.",
		})
		return
	}
	kubeClient, ok := kubeClientIface.(*kube.KubeClient)
	if !ok || kubeClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Kubernetes client"})
		return
	}
	// For audit/logging only: get cluster name from kubeClient or config (first cluster or unknown)
	clusterName := "unknown"
	cfg := config.Load("config/default.yaml")
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
	allNamespaces := false
	for i := 0; i < len(parts); i++ {
		if parts[i] == "-n" && i+1 < len(parts) {
			namespace = parts[i+1]
		}
		if parts[i] == "-A" || parts[i] == "--all-namespaces" {
			allNamespaces = true
		}
	}
	if allNamespaces {
		namespace = ""
	}
	output := ""
	success := true
	var opErr error
	var dryRun bool = req.DryRun
	switch verb {
	case "logs":
		if resource != "" {
			podName := resource
			podLogOpts := &corev1.PodLogOptions{
				Follow: false, // Set to true if you want to stream logs
			}
			req := kubeClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
			podLogs, err := req.Stream(c)
			if err != nil {
				success = false
				output = fmt.Sprintf("Error fetching logs: %v", err)
			} else {
				defer podLogs.Close()
				logBuf := new(bytes.Buffer)
				io.Copy(logBuf, podLogs)
				output = logBuf.String()
			}
		} else {
			success = false
			output = "Pod name not specified for logs command."
		}
		break
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
				if allNamespaces {
					output = "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE\n"
					for _, p := range pods.Items {
						output += p.Namespace + "\t"
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
		} else if resource == "cm" || resource == "configmaps" {
			configMaps, err := kubeClient.Clientset.CoreV1().ConfigMaps(namespace).List(c, metav1.ListOptions{})
			if err != nil {
				success = false
				output = "Error: " + err.Error()
			} else {
				if allNamespaces {
					output = "NAMESPACE\tNAME\tAGE\n"
					for _, cm := range configMaps.Items {
						output += cm.Namespace + "\t" + cm.Name + "\t" + cm.CreationTimestamp.Time.Format("2m") + "\n"
					}
				} else {
					output = "NAME\tAGE\n"
					for _, cm := range configMaps.Items {
						output += cm.Name + "\t" + cm.CreationTimestamp.Time.Format("2m") + "\n"
					}
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
	// Return real metrics if possible
	kubeClientIface, exists := c.Get("kubeClient")
	var podCount int
	var cpuUsage, memoryUsage string = "N/A", "N/A"
	if exists && kubeClientIface != nil {
		kubeClient, ok := kubeClientIface.(*kube.KubeClient)
		if ok && kubeClient != nil {
			podList, err := kubeClient.Clientset.CoreV1().Pods("").List(c, metav1.ListOptions{})
			if err == nil {
				podCount = len(podList.Items)
			}
			// --- Try to get node metrics from metrics-server ---
			// Import: k8s.io/metrics/pkg/client/clientset/versioned
			metricsClient, err := metricsclient.NewForConfig(kubeClient.Config)
			if err == nil {
				nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(c, metav1.ListOptions{})
				if err == nil && len(nodeMetricsList.Items) > 0 {
					// Aggregate CPU/memory usage across nodes
					var totalCPU, totalMem int64
					for _, nm := range nodeMetricsList.Items {
						cpu := nm.Usage.Cpu().MilliValue() // millicores
						mem := nm.Usage.Memory().Value()   // bytes
						totalCPU += cpu
						totalMem += mem
					}
					cpuUsage = fmt.Sprintf("%dm", totalCPU)
					memoryUsage = fmt.Sprintf("%dMi", totalMem/1024/1024)
				}
			}
		}
	}
	requestID := c.GetString("requestID")
	c.JSON(http.StatusOK, gin.H{
		"pods": podCount,
		"cpu":   cpuUsage,
		"memory": memoryUsage,
		"health": "OK",
		"request_id": requestID,
	})
}

func insightsHandler(c *gin.Context) {
	// Return cluster insights (stub, but structure for real impl)
	kubeClientIface, exists := c.Get("kubeClient")
	podWarnings := make([]gin.H, 0)
	if exists && kubeClientIface != nil {
		kubeClient, ok := kubeClientIface.(*kube.KubeClient)
		if ok && kubeClient != nil {
			podList, err := kubeClient.Clientset.CoreV1().Pods("").List(c, metav1.ListOptions{})
			if err == nil {
				for _, pod := range podList.Items {
					for _, cs := range pod.Status.ContainerStatuses {
						if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
							// Dynamic severity based on RestartCount
							sev := "high"
							if cs.RestartCount < 5 && cs.RestartCount >= 2 {
								sev = "medium"
							} else if cs.RestartCount == 1 {
								sev = "low"
							}
							podWarnings = append(podWarnings, gin.H{
								"type": "CrashLoopBackOff",
								"namespace": pod.Namespace,
								"pod": pod.Name,
								"container": cs.Name,
								"message": fmt.Sprintf("Pod %s (container %s, ns %s) is restarting frequently (restarts: %d)", pod.Name, cs.Name, pod.Namespace, cs.RestartCount),
								"severity": sev,
								"suggestion": fmt.Sprintf("Check logs with: kubectl logs %s -n %s -c %s", pod.Name, pod.Namespace, cs.Name),
							})
							// Do not break: allow multiple containers per pod
						}
					}
				}
			}
		}
	}
	// Filtering by severity
	severity := c.Query("severity")
	if severity != "" {
		filtered := make([]gin.H, 0, len(podWarnings))
		for _, w := range podWarnings {
			if val, ok := w["severity"]; ok && val == severity {
				filtered = append(filtered, w)
			}
		}
		podWarnings = filtered
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
