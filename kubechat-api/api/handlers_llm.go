package api

import (
	"net/http"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
)


// llmParseHandler: Calls local Ollama LLM to translate NL to kubectl/YAML/explanation
func llmParseHandler(c *gin.Context) {

	var req struct {
		Query string `json:"query" binding:"required,min=3,max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid query"})
		return
	}

	ollamaURL := "http://localhost:11434/api/generate"
	model := "llama3.2"
	prompt := `You are an expert Kubernetes assistant. Given a user request, output ONLY the most relevant kubectl command as JSON with the key: kubectl_command. 
If the user is asking a 'why' or troubleshooting question, output a diagnostic kubectl command (such as 'kubectl describe', 'kubectl get events', or 'kubectl logs') that would help investigate the issue. 
If you cannot generate a command, return an empty string for kubectl_command. Do NOT include YAML or explanations.

User request: ` + req.Query

	ollamaReq := map[string]interface{}{
		"model": model,
		"prompt": prompt,
		"stream": false,
	}
	// Optional: Log the prompt for debugging
	println("[DEBUG] LLM prompt:\n" + prompt)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(ollamaReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode request"})
		return
	}
	resp, err := http.Post(ollamaURL, "application/json", buf)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to contact LLM service", "details": err.Error()})
		return
	}
	defer resp.Body.Close()
	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to parse LLM response", "details": err.Error()})
		return
	}

	// Try to extract JSON from LLM response, even if inside code fences
	llmRaw := ollamaResp.Response
	// Optional: Log the raw LLM response for debugging
	println("[DEBUG] LLM raw response:\n" + llmRaw)
	jsonBlock := ""
	// Use regex to extract JSON block
	jsonRe := regexp.MustCompile("(?s)```json\\s*({.*?})\\s*```|({.*?})")
	matches := jsonRe.FindStringSubmatch(llmRaw)
	if len(matches) > 1 && matches[1] != "" {
		jsonBlock = matches[1]
	} else if len(matches) > 2 && matches[2] != "" {
		jsonBlock = matches[2]
	}
	if jsonBlock != "" {
		var llmOut map[string]interface{}
		if err := json.Unmarshal([]byte(jsonBlock), &llmOut); err == nil {
			// Only return the kubectl_command field if present
			if cmd, ok := llmOut["kubectl_command"].(string); ok {
				// Clean up placeholder namespaces and extra whitespace
				cleanCmd := regexp.MustCompile(`-n[ =]+(<|\[|\{)?namespace(\}|\]|>)?`).ReplaceAllString(cmd, "")
				cleanCmd = regexp.MustCompile(`['"“”‘’]`).ReplaceAllString(cleanCmd, "") // Remove quotes
				cleanCmd = regexp.MustCompile(`\s+`).ReplaceAllString(cleanCmd, " ") // Normalize whitespace
				cleanCmd = strings.TrimSpace(cleanCmd)
				c.JSON(http.StatusOK, gin.H{"kubectl_command": cleanCmd})
				return
			}
		}
	}
	// If not valid JSON or kubectl_command missing, try to extract a kubectl command from the raw response
	kubectlRe := regexp.MustCompile(`kubectl[^"\r]*`)
	cmdMatch := kubectlRe.FindString(llmRaw)
	if cmdMatch != "" {
		// Clean up placeholder namespaces and extra whitespace
		cleanCmd := regexp.MustCompile(`-n[ =]+(<|\[|\{)?namespace(\}|\]|>)?`).ReplaceAllString(cmdMatch, "")
		cleanCmd = regexp.MustCompile(`['"“”‘’]`).ReplaceAllString(cleanCmd, "") // Remove quotes
		cleanCmd = regexp.MustCompile(`\s+`).ReplaceAllString(cleanCmd, " ") // Normalize whitespace
		cleanCmd = strings.TrimSpace(cleanCmd)
		c.JSON(http.StatusOK, gin.H{"kubectl_command": cleanCmd})
		return
	}
	// If still nothing, return raw text
	c.JSON(http.StatusOK, gin.H{"llm_raw": ollamaResp.Response})
	return
}
