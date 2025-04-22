# KubeChat

KubeChat is a conversational Kubernetes assistant with a web-based frontend and a Go-based API backend.

## Project Structure

- `frontend/` — React-based web app for interacting with KubeChat. See [frontend/README.md](./frontend/README.md) for setup and usage.
- `kubechat-api/` — Go-based REST API server that handles Kubernetes commands and natural language queries. See [kubechat-api/README.md](./kubechat-api/README.md) for setup and API documentation.

## Getting Started

### 1. Running the API Backend
See [kubechat-api/README.md](./kubechat-api/README.md) for detailed instructions on:
- Prerequisites (Go, Kubernetes access, etc.)
- How to configure and run the API
- API endpoints and usage

### 2. Running the Frontend
See [frontend/README.md](./frontend/README.md) for detailed instructions on:
- Prerequisites (Node.js, npm, etc.)
- How to start the frontend app
- Environment configuration

## Which should I run?
- **To use KubeChat end-to-end:** Run both the API and frontend as described above.
- **To develop or debug only the API:** Follow [kubechat-api/README.md](./kubechat-api/README.md).
- **To work on the UI only:** Follow [frontend/README.md](./frontend/README.md).

## Contributing
Pull requests and issues are welcome! Please see the respective `README.md` in each directory for contribution guidelines and local development tips.

## License
MIT License. See [LICENSE](./LICENSE) for details.
