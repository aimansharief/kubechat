# KubeChat: Natural Language Interface for Kubernetes Management

KubeChat provides a conversational interface that translates natural language queries into Kubernetes operations, making cluster management accessible to team members with varying technical expertise while providing contextual insights and recommendations.

## Features

- **Chat Interface** - Clean, modern chat UI with message history, command suggestions, and confirmation prompts for potentially destructive operations
- **Natural Language Processing** - Translates plain English requests like "scale the frontend to 5 replicas" into proper kubectl commands
- **Context-Aware Insights** - Proactively identifies issues (e.g., CrashLoopBackOff pods) and suggests remediation steps with plain-language explanations
- **Command Validation** - Shows the actual kubectl command that will be executed with a dry-run option for verification before execution
- **Integrated Mini-Dashboard** - Compact visualization of critical cluster metrics with AI-generated health insights alongside the chat interface

## Prerequisites

- Node.js (v18 or higher)
- npm (v8 or higher)
- A modern web browser (Chrome, Firefox, Safari, or Edge)
- Kubernetes cluster access (for actual command execution)

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd kubechat

# Install dependencies
npm install
```

## Running the Application

### Development Mode

```bash
# Start the development server
npm run dev
```

The application will be available at http://localhost:5173

### Production Build

```bash
# Build the application
npm run build

# Preview the production build
npm run preview
```

## Project Structure

- `src/components/ChatInterface.tsx` - Main chat interface component
- `src/components/CommandPreview.tsx` - Component for displaying and validating kubectl commands
- `src/components/MiniDashboard.tsx` - Compact dashboard for cluster metrics and health insights

## License

This project is licensed under the MIT License - see the LICENSE file for details.
