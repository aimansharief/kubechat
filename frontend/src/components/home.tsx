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

  // Fix: Add missing state for input, suggestions
  const [inputValue, setInputValue] = useState("");
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [currentCommand, setCurrentCommand] = useState<Command | null>(null);
  const [showCommandPreview, setShowCommandPreview] = useState(false);
  const [pendingQuery, setPendingQuery] = useState<string>("");

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

    // Store the pending query for preview
    setPendingQuery(content);
    // Process the message (in a real app, this would involve NLP)
    processUserMessage(content);
  };

   const processUserMessage = async (content: string) => {
    try {
      const resp = await fetch("/api/v1/llm-parse", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query: content }),
      });
      const data = await resp.json();
      if (data.kubectl_command) {
        const command: Command = {
          naturalLanguage: content,
          kubectlCommand: data.kubectl_command,
          isDestructive: /delete|remove|scale|patch|apply/.test(data.kubectl_command),
        };
        setCurrentCommand(command);
        setShowCommandPreview(true); // Only show preview after LLM responds
      } else {
        setMessages((prev) => [
          ...prev,
          {
            id: Date.now().toString(),
            content: "Sorry, I could not generate a kubectl command for that query.",
            sender: "system",
            timestamp: new Date(),
          },
        ]);
      }
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          content: "There was an error contacting the backend.",
          sender: "system",
          timestamp: new Date(),
        },
      ]);
    }
  };

  const handleExecuteCommand = async (dryRun: boolean = false) => {
    if (!currentCommand) return;
    try {
      const resp = await fetch("/api/v1/execute", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          command: currentCommand.kubectlCommand,
          dry_run: dryRun,
        }),
      });
      const data = await resp.json();
      const result =
        data.result || data.output || JSON.stringify(data, null, 2);
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          content: result,
          sender: "system",
          timestamp: new Date(),
          isCommand: true,
          commandResult: result,
        },
      ]);
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          content: "There was an error executing the command.",
          sender: "system",
          timestamp: new Date(),
        },
      ]);
    }
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
            inputValue={inputValue}
            setInputValue={setInputValue}
            showSuggestions={showSuggestions}
            suggestions={suggestions}
            showCommandPreview={showCommandPreview}
            currentCommand={currentCommand}
            onExecuteCommand={() => handleExecuteCommand(false)}
            onDryRun={() => handleExecuteCommand(true)}
            onCancelCommand={handleCancelCommand}
            originalQuery={pendingQuery}
          />
        </div>
      </div>
    </div>
  );
}
