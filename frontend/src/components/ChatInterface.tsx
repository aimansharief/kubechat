import React, { useState, useRef, useEffect } from "react";
import { Send, AlertCircle, Terminal, Play, Info, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

// Types for props
interface Command {
  naturalLanguage: string;
  kubectlCommand: string;
  isDestructive: boolean;
}

interface Message {
  id: string;
  content: string;
  sender: "user" | "system";
  timestamp: Date;
  type?: "message" | "command" | "result" | "error" | "warning";
  command?: string;
  result?: string;
}

interface ChatInterfaceProps {
  messages: Message[];
  onSendMessage: (msg: string) => void;
  inputValue: string;
  setInputValue: (val: string) => void;
  showSuggestions?: boolean;
  suggestions?: string[];
  showCommandPreview: boolean;
  currentCommand: Command | null;
  onExecuteCommand: () => void;
  onDryRun: () => void;
  onCancelCommand: () => void;
  originalQuery?: string;
}

interface CommandPreviewProps {
  command: string;
  originalQuery: string;
  onExecute: () => void;
  onDryRun: () => void;
  onCancel: () => void;
}

const CommandPreview: React.FC<CommandPreviewProps> = ({
  command,
  originalQuery,
  onExecute,
  onDryRun,
  onCancel,
}) => {
  return (
    <div className="p-4 border-t border-b bg-slate-100">
      <div className="flex justify-between items-center mb-2">
        <h3 className="text-sm font-medium">Command Preview</h3>
        <Button variant="ghost" size="sm" onClick={onCancel}>
          <X className="h-4 w-4" />
        </Button>
      </div>
      <div className="mb-2">
        <div className="text-xs text-slate-500">Original query:</div>
        <div className="text-sm">{originalQuery}</div>
      </div>
      <div className="mb-3">
        <div className="text-xs text-slate-500">Generated command:</div>
        <div className="bg-slate-800 text-white p-2 rounded font-mono text-sm">
          {command}
        </div>
      </div>
      <div className="flex space-x-2 justify-end">
        <Button variant="outline" size="sm" onClick={onCancel}>
          Cancel
        </Button>
        <Button variant="outline" size="sm" onClick={onDryRun}>
          <Play className="h-4 w-4 mr-1" />
          Dry Run
        </Button>
        <Button size="sm" onClick={onExecute}>
          <Terminal className="h-4 w-4 mr-1" />
          Execute
        </Button>
      </div>
    </div>
  );
};



const ChatInterface: React.FC<ChatInterfaceProps> = ({
  messages,
  onSendMessage,
  inputValue,
  setInputValue,
  showSuggestions = false,
  suggestions = [],
  showCommandPreview,
  currentCommand,
  onExecuteCommand,
  onDryRun,
  onCancelCommand,
  originalQuery = "",
  isConnected = true,
}) => {
    const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
  };

  const handleSendMessage = () => {
    if (inputValue.trim() === "") return;
    onSendMessage(inputValue);
    setInputValue("");
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handleSendMessage();
    }
  };

  const handleSuggestionClick = (suggestion: string) => {
    setInputValue(suggestion);
    inputRef.current?.focus();
  };





  const getMessageStyle = (type?: string) => {
    switch (type) {
      case "command":
        return "bg-slate-800 text-white font-mono";
      case "result":
        return "bg-slate-700 text-white font-mono";
      case "error":
        return "bg-red-100 text-red-800 border-l-4 border-red-500";
      case "warning":
        return "bg-amber-100 text-amber-800 border-l-4 border-amber-500";
      default:
        return "bg-white";
    }
  };

  const getMessageIcon = (type?: string) => {
    switch (type) {
      case "command":
        return <Terminal className="h-4 w-4 mr-2" />;
      case "error":
        return <AlertCircle className="h-4 w-4 mr-2 text-red-500" />;
      case "warning":
        return <Info className="h-4 w-4 mr-2 text-amber-500" />;
      default:
        return null;
    }
  };

  return (
    <div className="flex flex-col h-full bg-slate-50 rounded-lg overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b bg-white">
        <div className="flex items-center">
          <h2 className="text-xl font-semibold">KubeChat</h2>
          <Badge
            variant={isConnected ? "default" : "destructive"}
            className="ml-3"
          >
            {isConnected ? "Connected" : "Disconnected"}
          </Badge>
        </div>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="outline" size="sm">
                <Terminal className="h-4 w-4 mr-2" />
                Command History
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>View your command history</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>

      {/* Messages Area */}
      <ScrollArea className="flex-1 p-4">
        <div className="space-y-4">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.sender === "user" ? "justify-end" : "justify-start"}`}
            >
              <div
                className={`max-w-3/4 rounded-lg p-3 ${message.sender === "user" ? "bg-primary text-primary-foreground" : getMessageStyle(message.type)}`}
              >
                {message.sender === "system" && message.type !== "message" && (
                  <div className="flex items-center mb-1 text-xs font-medium">
                    {getMessageIcon(message.type)}
                    {message.type === "command"
                      ? "Command"
                      : message.type === "result"
                        ? "Result"
                        : message.type === "error"
                          ? "Error"
                          : "Warning"}
                  </div>
                )}
                <div className="whitespace-pre-wrap">{message.content}</div>
              </div>
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>
      </ScrollArea>

      {/* Command Preview */}
      {showCommandPreview && currentCommand && (
        <CommandPreview
          command={currentCommand.kubectlCommand}
          originalQuery={currentCommand.naturalLanguage}
          onExecute={onExecuteCommand}
          onDryRun={onDryRun}
          onCancel={onCancelCommand}
        />
      )}

      {/* Input Area */}
      <div className="p-4 border-t bg-white">
        {showSuggestions && (
          <div className="mb-3 flex flex-wrap gap-2">
            {suggestions.map((suggestion, index) => (
              <Badge
                key={index}
                variant="outline"
                className="cursor-pointer hover:bg-slate-100"
                onClick={() => handleSuggestionClick(suggestion)}
              >
                {suggestion}
              </Badge>
            ))}
          </div>
        )}
        <div className="flex items-center space-x-2">
          <Avatar className="h-8 w-8">
            <AvatarImage
              src="https://api.dicebear.com/7.x/avataaars/svg?seed=user"
              alt="User"
            />
            <AvatarFallback>U</AvatarFallback>
          </Avatar>
          <Input
            ref={inputRef}
            value={inputValue}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder="Type your Kubernetes query or command..."
            className="flex-1"
          />
          <Button
            onClick={handleSendMessage}
            disabled={inputValue.trim() === ""}
          >
            <Send className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
};

export default ChatInterface;
