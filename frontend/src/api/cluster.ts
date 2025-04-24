// API utilities for cluster metrics, pods, and insights
export interface ClusterMetric {
  name: string;
  value: number;
  max: number;
  unit: string;
}

export interface PodStatus {
  name: string;
  namespace: string;
  status: string;
  restarts: number;
  age: string;
}

export interface ClusterInsight {
  type: string;
  message: string;
  timestamp: string;
}

export async function fetchClusterMetrics(): Promise<ClusterMetric[]> {
  const resp = await fetch("/api/v1/metrics");
  if (!resp.ok) throw new Error("Failed to fetch metrics");
  return resp.json();
}

export async function fetchPodStatuses(): Promise<PodStatus[]> {
  const resp = await fetch("/api/v1/pods");
  if (!resp.ok) throw new Error("Failed to fetch pods");
  return resp.json();
}

export async function fetchClusterInsights(): Promise<ClusterInsight[]> {
  const resp = await fetch("/api/v1/insights");
  if (!resp.ok) throw new Error("Failed to fetch insights");
  return resp.json();
}

