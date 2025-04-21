import React, { useState } from "react";
import ChatInterface from "./ChatInterface";
import MiniDashboard from "./MiniDashboard";
import CommandPreview from "./CommandPreview";
import { Separator } from "./ui/separator";

interface Message {
  id: string;
  content: string;
  sender: "user" | "system";
  timestamp: Date;
  isCommand?: boolean;
  commandResult?: string;
}

interface Command {
  naturalLanguage: string;
  kubectlCommand: string;
  isDestructive: boolean;
}

export default function Home() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "1",
      content:
        "Welcome to KubeChat! How can I help you manage your Kubernetes cluster today?",
      sender: "system",
      timestamp: new Date(),
    },
  ]);

  const [currentCommand, setCurrentCommand] = useState<Command | null>(null);
  const [showCommandPreview, setShowCommandPreview] = useState(false);

  // Mock cluster metrics for demonstration
  const clusterMetrics = {
    cpu: {
      usage: 45,
      total: 100,
    },
    memory: {
      usage: 60,
      total: 128,
    },
    pods: {
      running: 12,
      pending: 2,
      failed: 1,
      total: 15,
    },
    nodes: {
      ready: 3,
      notReady: 0,
      total: 3,
    },
    alerts: [
      {
        id: "alert-1",
        severity: "warning",
        message: "High memory usage in frontend deployment",
        timestamp: new Date(),
      },
      {
        id: "alert-2",
        severity: "critical",
        message: "CrashLoopBackOff in auth-service pod",
        timestamp: new Date(),
      },
    ],
  };

  const handleSendMessage = (content: string) => {
    // Add user message to chat
    const userMessage: Message = {
      id: Date.now().toString(),
      content,
      sender: "user",
      timestamp: new Date(),
    };
    setMessages((prev) => [...prev, userMessage]);

    // Process the message (in a real app, this would involve NLP)
    processUserMessage(content);
  };

  const processUserMessage = (content: string) => {
    // Mock NLP processing - in a real app this would use an actual NLP service
    setTimeout(() => {
      // Example command translation
      if (
        content.toLowerCase().includes("scale") &&
        content.toLowerCase().includes("frontend")
      ) {
        const command: Command = {
          naturalLanguage: content,
          kubectlCommand: "kubectl scale deployment frontend --replicas=5",
          isDestructive: false,
        };
        setCurrentCommand(command);
        setShowCommandPreview(true);
      } else if (
        content.toLowerCase().includes("delete") ||
        content.toLowerCase().includes("remove")
      ) {
        const command: Command = {
          naturalLanguage: content,
          kubectlCommand: "kubectl delete pod crashed-pod-abc123",
          isDestructive: true,
        };
        setCurrentCommand(command);
        setShowCommandPreview(true);
      } else {
        // Generic response for other queries
        const systemResponse: Message = {
          id: Date.now().toString(),
          content: `I'll help you with "${content}". What specific information are you looking for?`,
          sender: "system",
          timestamp: new Date(),
        };
        setMessages((prev) => [...prev, systemResponse]);
      }
    }, 1000);
  };

  const handleExecuteCommand = (dryRun: boolean = false) => {
    if (!currentCommand) return;

    // In a real app, this would execute the kubectl command
    const result = dryRun
      ? `[DRY RUN] Command would execute: ${currentCommand.kubectlCommand}`
      : `Executed: ${currentCommand.kubectlCommand}\n\nResult: Operation successful`;

    const systemResponse: Message = {
      id: Date.now().toString(),
      content: result,
      sender: "system",
      timestamp: new Date(),
      isCommand: true,
      commandResult: result,
    };

    setMessages((prev) => [...prev, systemResponse]);
    setShowCommandPreview(false);
    setCurrentCommand(null);
  };

  const handleCancelCommand = () => {
    setShowCommandPreview(false);
    setCurrentCommand(null);

    const systemResponse: Message = {
      id: Date.now().toString(),
      content: "Command cancelled. How else can I help you?",
      sender: "system",
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, systemResponse]);
  };

  return (
    <div className="flex h-screen w-full bg-background">
      {/* Mini Dashboard */}
      <div className="w-1/3 border-r border-border">
        <MiniDashboard metrics={clusterMetrics} />
      </div>

      {/* Chat Interface */}
      <div className="flex flex-col w-2/3">
        <div className="flex-1 overflow-hidden">
          <ChatInterface
            messages={messages}
            onSendMessage={handleSendMessage}
          />
        </div>

        {/* Command Preview */}
        {showCommandPreview && currentCommand && (
          <div className="border-t border-border p-4">
            <CommandPreview
              command={currentCommand}
              onExecute={handleExecuteCommand}
              onDryRun={() => handleExecuteCommand(true)}
              onCancel={handleCancelCommand}
            />
          </div>
        )}
      </div>
    </div>
  );
}
