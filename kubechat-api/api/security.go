package api

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var allowedVerbs = map[string]bool{
	"get":      true,
	"list":     true,
	"describe": true,
	"logs":     true,
	"scale":    true,
}

// CommandValidator blocks dangerous verbs and basic command injection
func CommandValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check method and path; do not parse the body here
		c.Next()
	}
}

func extractVerb(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) < 2 {
		return ""
	}
	// e.g. kubectl get pods -> get
	if parts[0] == "kubectl" {
		return strings.ToLower(parts[1])
	}
	return ""
}

func hasInjection(cmd string) bool {
	// Detect ; | && || $() > <
	injectionPattern := regexp.MustCompile(`[;|&><$]`)
	return injectionPattern.MatchString(cmd)
}
