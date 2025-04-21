import React from "react";
import {
  AlertCircle,
  CheckCircle2,
  Play,
  AlertTriangle,
  XCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface CommandPreviewProps {
  originalQuery: string;
  translatedCommand: string;
  isDestructive?: boolean;
  onExecute?: () => void;
  onModify?: () => void;
  onDryRun?: () => void;
  onCancel?: () => void;
}

const CommandPreview = ({
  originalQuery = "Scale the frontend deployment to 5 replicas",
  translatedCommand = "kubectl scale deployment frontend --replicas=5",
  isDestructive = false,
  onExecute = () => console.log("Execute command"),
  onModify = () => console.log("Modify command"),
  onDryRun = () => console.log("Dry run command"),
  onCancel = () => console.log("Cancel command"),
}: CommandPreviewProps) => {
  return (
    <Card className="w-full bg-white border-2 shadow-md">
      <CardHeader className="pb-2">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg font-medium">Command Preview</CardTitle>
          {isDestructive && (
            <Badge variant="destructive" className="flex items-center gap-1">
              <AlertCircle size={14} />
              Potentially Destructive
            </Badge>
          )}
        </div>
      </CardHeader>
      <Separator />
      <CardContent className="pt-4">
        <div className="space-y-3">
          <div>
            <p className="text-sm font-medium text-muted-foreground">
              Original Query:
            </p>
            <p className="text-sm mt-1">{originalQuery}</p>
          </div>
          <div>
            <div className="flex items-center gap-2">
              <p className="text-sm font-medium text-muted-foreground">
                Translated Command:
              </p>
              {isDestructive ? (
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <AlertTriangle size={16} className="text-amber-500" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>This command may modify or delete resources</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              ) : (
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <CheckCircle2 size={16} className="text-green-500" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>This command is safe to execute</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
            </div>
            <div className="mt-1 p-2 bg-slate-100 rounded-md font-mono text-sm overflow-x-auto">
              {translatedCommand}
            </div>
          </div>
        </div>
      </CardContent>
      <CardFooter className="flex justify-end gap-2 pt-2">
        <Button variant="outline" size="sm" onClick={onCancel}>
          <XCircle className="mr-1 h-4 w-4" />
          Cancel
        </Button>
        <Button variant="outline" size="sm" onClick={onModify}>
          Modify
        </Button>
        <Button variant="secondary" size="sm" onClick={onDryRun}>
          Dry Run
        </Button>
        <Button
          variant={isDestructive ? "destructive" : "default"}
          size="sm"
          onClick={onExecute}
        >
          <Play className="mr-1 h-4 w-4" />
          Execute
        </Button>
      </CardFooter>
    </Card>
  );
};

export default CommandPreview;
