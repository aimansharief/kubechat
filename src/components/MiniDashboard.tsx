import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  AlertCircle,
  CheckCircle,
  Info,
  Server,
  Cpu,
  HardDrive,
  Database,
  Activity,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface ClusterMetric {
  name: string;
  value: number;
  max: number;
  unit: string;
}

interface PodStatus {
  name: string;
  namespace: string;
  status: "Running" | "Pending" | "Failed" | "CrashLoopBackOff" | "Completed";
  restarts: number;
  age: string;
}

interface ClusterInsight {
  type: "info" | "warning" | "error" | "success";
  message: string;
  timestamp: string;
}

interface MiniDashboardProps {
  onInsightClick?: (insight: ClusterInsight) => void;
}

const MiniDashboard: React.FC<MiniDashboardProps> = ({
  onInsightClick = () => {},
}) => {
  const [activeTab, setActiveTab] = useState("overview");
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Mock data - in a real app, this would come from API calls to the Kubernetes cluster
  const clusterMetrics: ClusterMetric[] = [
    { name: "CPU Usage", value: 45, max: 100, unit: "%" },
    { name: "Memory Usage", value: 62, max: 100, unit: "%" },
    { name: "Disk Usage", value: 38, max: 100, unit: "%" },
    { name: "Network I/O", value: 25, max: 100, unit: "MB/s" },
  ];

  const podStatuses: PodStatus[] = [
    {
      name: "frontend-deployment-7b9f6d5b8c-1a2b3",
      namespace: "default",
      status: "Running",
      restarts: 0,
      age: "2d",
    },
    {
      name: "backend-api-5c7d8e9f4a-4c5d6",
      namespace: "default",
      status: "Running",
      restarts: 0,
      age: "2d",
    },
    {
      name: "database-0",
      namespace: "db",
      status: "Running",
      restarts: 0,
      age: "5d",
    },
    {
      name: "cache-redis-6b7c8d9e0f-7e8f9",
      namespace: "cache",
      status: "Running",
      restarts: 0,
      age: "3d",
    },
    {
      name: "monitoring-prometheus-1a2b3c4d5e-6f7g8",
      namespace: "monitoring",
      status: "Running",
      restarts: 0,
      age: "7d",
    },
    {
      name: "logging-fluentd-9h0i1j2k3l-4m5n6",
      namespace: "logging",
      status: "CrashLoopBackOff",
      restarts: 5,
      age: "1d",
    },
  ];

  const clusterInsights: ClusterInsight[] = [
    {
      type: "error",
      message:
        "Pod logging-fluentd-9h0i1j2k3l-4m5n6 is in CrashLoopBackOff state with 5 restarts",
      timestamp: "10 minutes ago",
    },
    {
      type: "warning",
      message:
        "Memory usage trending high (62%) - consider scaling up resources",
      timestamp: "25 minutes ago",
    },
    {
      type: "info",
      message:
        "New deployment detected: frontend-deployment updated to version v1.2.3",
      timestamp: "1 hour ago",
    },
    {
      type: "success",
      message: "All critical services are running optimally",
      timestamp: "2 hours ago",
    },
  ];

  const handleRefresh = () => {
    setIsRefreshing(true);
    // Simulate refresh delay
    setTimeout(() => {
      setIsRefreshing(false);
    }, 1000);
  };

  const getStatusColor = (status: PodStatus["status"]) => {
    switch (status) {
      case "Running":
        return "bg-green-500";
      case "Pending":
        return "bg-yellow-500";
      case "Failed":
        return "bg-red-500";
      case "CrashLoopBackOff":
        return "bg-red-500";
      case "Completed":
        return "bg-blue-500";
      default:
        return "bg-gray-500";
    }
  };

  const getInsightIcon = (type: ClusterInsight["type"]) => {
    switch (type) {
      case "error":
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      case "warning":
        return <AlertCircle className="h-4 w-4 text-yellow-500" />;
      case "info":
        return <Info className="h-4 w-4 text-blue-500" />;
      case "success":
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      default:
        return <Info className="h-4 w-4" />;
    }
  };

  return (
    <div className="h-full w-full flex flex-col bg-background border-r">
      <div className="p-4 flex justify-between items-center border-b">
        <div className="flex items-center space-x-2">
          <Server className="h-5 w-5" />
          <h2 className="text-lg font-semibold">Cluster Dashboard</h2>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleRefresh}
          disabled={isRefreshing}
          className="h-8 w-8 p-0"
        >
          <RefreshCw
            className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
          />
        </Button>
      </div>

      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="flex-1 flex flex-col"
      >
        <div className="px-4 pt-2">
          <TabsList className="w-full">
            <TabsTrigger value="overview" className="flex-1">
              Overview
            </TabsTrigger>
            <TabsTrigger value="pods" className="flex-1">
              Pods
            </TabsTrigger>
            <TabsTrigger value="insights" className="flex-1">
              Insights
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-hidden">
          <TabsContent value="overview" className="h-full p-4 space-y-4 mt-0">
            <div className="grid grid-cols-2 gap-4">
              {clusterMetrics.map((metric, index) => (
                <Card key={index} className="overflow-hidden">
                  <CardHeader className="p-3">
                    <CardTitle className="text-sm font-medium flex items-center">
                      {metric.name === "CPU Usage" && (
                        <Cpu className="h-4 w-4 mr-2" />
                      )}
                      {metric.name === "Memory Usage" && (
                        <HardDrive className="h-4 w-4 mr-2" />
                      )}
                      {metric.name === "Disk Usage" && (
                        <Database className="h-4 w-4 mr-2" />
                      )}
                      {metric.name === "Network I/O" && (
                        <Activity className="h-4 w-4 mr-2" />
                      )}
                      {metric.name}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="p-3 pt-0">
                    <div className="text-2xl font-bold mb-1">
                      {metric.value}
                      {metric.unit}
                    </div>
                    <Progress value={metric.value} className="h-2" />
                  </CardContent>
                </Card>
              ))}
            </div>

            <Card>
              <CardHeader className="p-3">
                <CardTitle className="text-sm font-medium">
                  Pod Status Summary
                </CardTitle>
              </CardHeader>
              <CardContent className="p-3 pt-0">
                <div className="flex space-x-2">
                  <Badge variant="outline" className="flex items-center gap-1">
                    <div className="h-2 w-2 rounded-full bg-green-500"></div>
                    <span>
                      Running:{" "}
                      {podStatuses.filter((p) => p.status === "Running").length}
                    </span>
                  </Badge>
                  <Badge variant="outline" className="flex items-center gap-1">
                    <div className="h-2 w-2 rounded-full bg-yellow-500"></div>
                    <span>
                      Pending:{" "}
                      {podStatuses.filter((p) => p.status === "Pending").length}
                    </span>
                  </Badge>
                  <Badge variant="outline" className="flex items-center gap-1">
                    <div className="h-2 w-2 rounded-full bg-red-500"></div>
                    <span>
                      Failed/CrashLoop:{" "}
                      {
                        podStatuses.filter((p) =>
                          ["Failed", "CrashLoopBackOff"].includes(p.status),
                        ).length
                      }
                    </span>
                  </Badge>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="p-3">
                <CardTitle className="text-sm font-medium">
                  Latest Insights
                </CardTitle>
              </CardHeader>
              <CardContent className="p-3 pt-0 space-y-2">
                {clusterInsights.slice(0, 2).map((insight, index) => (
                  <div
                    key={index}
                    className="flex items-start space-x-2 text-sm p-2 rounded-md hover:bg-muted cursor-pointer"
                    onClick={() => onInsightClick(insight)}
                  >
                    {getInsightIcon(insight.type)}
                    <div className="flex-1">
                      <p>{insight.message}</p>
                      <p className="text-xs text-muted-foreground">
                        {insight.timestamp}
                      </p>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="pods" className="h-full p-4 mt-0">
            <Card className="h-full">
              <CardHeader className="p-3">
                <CardTitle className="text-sm font-medium">
                  Pod Status
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <ScrollArea className="h-[calc(100vh-220px)]">
                  <div className="p-3">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="text-left text-muted-foreground">
                          <th className="pb-2">Status</th>
                          <th className="pb-2">Name</th>
                          <th className="pb-2">Namespace</th>
                          <th className="pb-2">Restarts</th>
                          <th className="pb-2">Age</th>
                        </tr>
                      </thead>
                      <tbody>
                        {podStatuses.map((pod, index) => (
                          <tr key={index} className="border-t border-border">
                            <td className="py-2">
                              <TooltipProvider>
                                <Tooltip>
                                  <TooltipTrigger>
                                    <div
                                      className={`h-3 w-3 rounded-full ${getStatusColor(pod.status)}`}
                                    ></div>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <p>{pod.status}</p>
                                  </TooltipContent>
                                </Tooltip>
                              </TooltipProvider>
                            </td>
                            <td
                              className="py-2 max-w-[150px] truncate"
                              title={pod.name}
                            >
                              {pod.name}
                            </td>
                            <td className="py-2">{pod.namespace}</td>
                            <td className="py-2">{pod.restarts}</td>
                            <td className="py-2">{pod.age}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="insights" className="h-full p-4 mt-0">
            <Card className="h-full">
              <CardHeader className="p-3">
                <CardTitle className="text-sm font-medium">
                  Cluster Insights
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <ScrollArea className="h-[calc(100vh-220px)]">
                  <div className="p-3 space-y-3">
                    {clusterInsights.map((insight, index) => (
                      <div
                        key={index}
                        className="flex items-start space-x-3 p-3 rounded-md border hover:bg-muted cursor-pointer"
                        onClick={() => onInsightClick(insight)}
                      >
                        {getInsightIcon(insight.type)}
                        <div className="flex-1">
                          <p>{insight.message}</p>
                          <p className="text-xs text-muted-foreground mt-1">
                            {insight.timestamp}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
};

export default MiniDashboard;
