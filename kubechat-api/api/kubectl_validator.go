package api

import (
	"bytes"
	"io"
	"fmt"
	"regexp"
	"strings"
	"time"
	"runtime/debug"

	"kubechat-api/config"
	"kubechat-api/kube"
	"kubechat-api/auditlog"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func KubectlValidator(cfg config.Config, kubeClient *kube.KubeClient, logger *zap.Logger) gin.HandlerFunc {
	allowedVerbs := map[string]bool{"get": true, "list": true, "describe": true, "scale": true}
	blockedVerbs := map[string]bool{"delete": true, "edit": true}
	// Allow all namespaces (optionally restrict by config later)
	// allowedNamespaces := map[string]bool{"default": true}
	// for _, ns := range []string{"default"} { // TODO: pull from config if available
	//     allowedNamespaces[ns] = true
	// }
	resourceNameRegex := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	return func(c *gin.Context) {
	fmt.Println("DEBUG: KubectlValidator middleware running - should allow all namespaces - ", time.Now())
		var req struct {
			Command string `json:"command"`
		}
		// Use cached body from context
		if v, exists := c.Get("rawBody"); exists {
			if bodyBytes, ok := v.([]byte); ok {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Next()
			return
		}
		parts := strings.Fields(req.Command)
		if len(parts) < 3 || parts[0] != "kubectl" {
			c.Next()
			return
		}
		// Re-cache the body for downstream handlers
		if v, exists := c.Get("rawBody"); exists {
			if bodyBytes, ok := v.([]byte); ok {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		verb := parts[1]
		resource := parts[2]
		var namespace string = "default"
		var resourceName string = ""
		for i := 0; i < len(parts); i++ {
			if parts[i] == "-n" && i+1 < len(parts) {
				namespace = parts[i+1]
			}
			if (parts[i] == resource || strings.HasPrefix(parts[i], resource+"/")) && i+1 < len(parts) {
				candidate := parts[i+1]
				if !strings.HasPrefix(candidate, "-") {
					resourceName = candidate
					fmt.Println("DEBUG: resourceName set to", resourceName)
				}
			}
		}
		fmt.Println("DEBUG: Final resourceName:", resourceName)
		if blockedVerbs[verb] {
			fmt.Println("DEBUG: logAndAbort called for blocked verb:", verb)
			fmt.Println("STACK TRACE:\n", string(debug.Stack()))
			logAndAbort(c, logger, req.Command, "Blocked dangerous verb", verb)
			return
		}
		if !allowedVerbs[verb] {
			fmt.Println("DEBUG: logAndAbort called for not allowed verb:", verb)
			fmt.Println("STACK TRACE:\n", string(debug.Stack()))
			logAndAbort(c, logger, req.Command, "Verb not allowed", verb)
			return
		}

		if resourceName != "" {
			if !resourceNameRegex.MatchString(resourceName) {
				fmt.Println("DEBUG: logAndAbort called for resource name not whitelisted:", resourceName)
				fmt.Println("STACK TRACE:\n", string(debug.Stack()))
				logAndAbort(c, logger, req.Command, "Resource name not whitelisted", resourceName)
				return
			}
		}
		// RBAC check (SelfSubjectAccessReview)
		ssar := &authorizationv1.SelfSubjectAccessReview{
			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Namespace: namespace,
					Verb:      verb,
					Resource:  resource,
					Name:      resourceName,
				},
			},
		}
		result, err := kubeClient.Clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(c, ssar, metav1.CreateOptions{})
		if err != nil || !result.Status.Allowed {
			msg := "RBAC denied"
			if err != nil {
				msg = err.Error()
			}
			fmt.Println("DEBUG: logAndAbort called for RBAC denial. Namespace:", namespace, "Verb:", verb, "Resource:", resource, "resourceName:", resourceName)
			fmt.Println("STACK TRACE:\n", string(debug.Stack()))
			logAndAbort(c, logger, req.Command, msg, "RBAC")
			return
		}
		c.Next()
	}
}

func logAndAbort(c *gin.Context, logger *zap.Logger, command, reason, detail string) {
	userID := c.GetString("userID")
	auditlog.AuditLog(logger, auditlog.AuditEntry{
		Timestamp: time.Now(),
		UserID:    userID,
		Cluster:   "dev-cluster",
		Command:   command,
		Success:   false,
		Details:   fmt.Sprintf("%s: %s", reason, detail),
	})
	c.AbortWithStatusJSON(403, gin.H{
		"error":   reason,
		"details": detail,
		"code":    "ERR_KUBECTL_VALIDATION",
	})
}
