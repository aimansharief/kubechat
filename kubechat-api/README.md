# KubeChat API

KubeChat API is a RESTful backend service for natural language-driven Kubernetes operations. It leverages a local LLM (Ollama) to translate user queries into actionable `kubectl` commands, and provides endpoints for parsing, executing, validating, and troubleshooting Kubernetes resources.

## Features
- **Natural Language to Kubectl**: Converts user queries to safe, actionable `kubectl` commands.
- **Command Execution**: Securely executes validated commands against your Kubernetes cluster.
- **Dry Run & Validation**: Validate and dry-run commands before actual execution.
- **Cluster Insights**: Get real-time metrics, health, and insights from your cluster.
- **Suggestions**: Get next-step suggestions for Kubernetes resources.
- **RBAC & Security**: Built-in command validation and basic RBAC enforcement.

## Requirements
- Go 1.18+
- Kubernetes cluster and valid kubeconfig
- [Ollama](https://ollama.com/) LLM running locally (default: `http://localhost:11434`)
- (Optional) Metrics-server installed in your cluster for metrics endpoints

> **⚠️ IMPORTANT: Ollama must be running locally for all LLM-powered endpoints (such as `/api/v1/llm-parse`) to work!**
> 
> Start Ollama with:
> ```sh
> ollama serve
> ```
> See [Ollama docs](https://ollama.com/) for details.

> **⚠️ IMPORTANT: kubeconfig.yaml must be placed in the root of this repo:**
> 
> `/Users/admin/Documents/workspace/kubechat/kubechat-api/kubeconfig.yaml`
> 
> This file is required for cluster access. If you use a different path, update the config accordingly.

## Setup
1. **Clone the repo**
   ```sh
   git clone <your-repo-url>
   cd kubechat-api
   ```
2. **Install Go dependencies**
   ```sh
   go mod tidy
   ```
3. **Start Ollama LLM**
   Ensure Ollama is running locally (see [Ollama docs](https://ollama.com/)).
4. **Configure kubeconfig**
   Place your kubeconfig at `kubeconfig.yaml` or update the path in config.
5. **Run the API**
   ```sh
   go run main.go
   ```
   The server will start at `http://localhost:8080`.

## API Endpoints
| Method | Endpoint               | Description                                  |
|--------|------------------------|----------------------------------------------|
| POST   | `/api/v1/llm-parse`    | Use LLM to parse NL query to kubectl         |
| POST   | `/api/v1/execute`      | Execute a validated kubectl command          |
| POST   | `/api/v1/dry-run`      | Validate a kubectl command (no execution)    |
| GET    | `/api/v1/context`      | Get namespaces and pod count                 |
| GET    | `/api/v1/insights`     | Get cluster insights                         |
| GET    | `/api/v1/metrics`      | Get cluster metrics                          |
| GET    | `/api/v1/health`       | Health check                                 |
| GET    | `/api/v1/cluster-health`| Cluster health summary                      |

See the provided Postman collection for request/response examples.

## Security Notes
- Only safe `kubectl` verbs are allowed (get, list, describe, logs, scale).
- Dangerous verbs (delete, edit, etc.) are blocked by default.
- Input validation and basic RBAC checks are enforced.

## Development & Testing
- All endpoints are tested via the included Postman collection (`kubechat-api.postman_collection.json`).
- Unit tests are in `api/handlers_test.go`.

## License
MIT
