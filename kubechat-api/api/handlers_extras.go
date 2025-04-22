package api

import (
	"bytes"
	"io"
	"github.com/gin-gonic/gin"
	"kubechat-api/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// dryRunHandler validates a command without executing it
func dryRunHandler(c *gin.Context) {
	// Validate the command using RBAC and syntax, but do not execute
	var req struct {
		Command string `json:"command" binding:"required,min=3,max=500"`
	}
	// Use cached body from context
	if v, exists := c.Get("rawBody"); exists {
		if bodyBytes, ok := v.([]byte); ok {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}
	if hasInjection(req.Command) {
		c.JSON(400, gin.H{"error": "Potential command injection detected", "details": req.Command})
		return
	}
	verb := extractVerb(req.Command)
	if verb == "" {
		c.JSON(400, gin.H{"error": "Could not extract verb from command"})
		return
	}
	// RBAC: Use the same logic as in KubectlValidator, but do not execute
	allowedVerbs := map[string]bool{"get": true, "list": true, "describe": true, "logs": true, "scale": true}
	if !allowedVerbs[verb] {
		c.JSON(403, gin.H{"result": "Command blocked by RBAC", "success": false})
		return
	}
	c.JSON(200, gin.H{"result": "Command validated successfully", "success": true})
}

// contextHandler returns current cluster state
func contextHandler(c *gin.Context) {
	// Retrieve kubeClient from gin.Context (set in main.go as a middleware or pass as closure)
	kubeClientIface, exists := c.Get("kubeClient")
	if !exists || kubeClientIface == nil {
		c.JSON(500, gin.H{"error": "Kubernetes client not available"})
		return
	}
	kubeClient, ok := kubeClientIface.(*kube.KubeClient)
	if !ok || kubeClient == nil {
		c.JSON(500, gin.H{"error": "Invalid Kubernetes client"})
		return
	}
	// List namespaces
	nsList, err := kubeClient.Clientset.CoreV1().Namespaces().List(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to list namespaces", "details": err.Error()})
		return
	}
	namespaces := []string{}
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}
	// List pods (all namespaces)
	podList, err := kubeClient.Clientset.CoreV1().Pods("").List(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to list pods", "details": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"namespaces": namespaces,
		"pods": len(podList.Items),
	})
}

// suggestionsHandler removed
