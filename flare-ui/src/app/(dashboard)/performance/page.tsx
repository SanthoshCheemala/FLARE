"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, Clock, Cpu, HardDrive, Zap, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/api-client";

interface PerformanceMetrics {
  total_time_seconds: number;
  total_time_formatted: string;
  key_gen_time_seconds: number;
  key_gen_time_formatted: string;
  key_gen_percent: number;
  hashing_time_seconds: number;
  hashing_time_formatted: string;
  hashing_percent: number;
  witness_time_seconds: number;
  witness_time_formatted: string;
  witness_percent: number;
  intersection_time_seconds: number;
  intersection_time_formatted: string;
  intersection_percent: number;
  num_workers: number;
  total_operations: number;
  throughput_ops_per_sec: number;
}

interface MemoryMetrics {
  alloc_mb: number;
  total_alloc_mb: number;
  sys_mb: number;
  num_gc: number;
  goroutines: number;
}

export default function PerformancePage() {
  const [perfMetrics, setPerfMetrics] = useState<PerformanceMetrics | null>(null);
  const [memMetrics, setMemMetrics] = useState<MemoryMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchMetrics = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getPerformanceMetrics();
      setPerfMetrics(data.performance);
      setMemMetrics(data.memory);
    } catch (err) {
      console.error("Failed to fetch performance metrics:", err);
      setError("Failed to load performance metrics. Please ensure the backend is running.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMetrics();
    // Auto-refresh every 10 seconds
    const interval = setInterval(fetchMetrics, 10000);
    return () => clearInterval(interval);
  }, []);

  if (loading && !perfMetrics) {
    return <div className="p-8">Loading performance data...</div>;
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-600">{error}</p>
          <Button onClick={fetchMetrics} className="mt-4" variant="outline">
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  if (!perfMetrics || !memMetrics) {
    return <div className="p-8">No performance data available yet. Run a screening to generate metrics.</div>;
  }

  const phases = [
    { name: "Key Generation", percent: perfMetrics.key_gen_percent, color: "from-blue-500 to-blue-600", time: perfMetrics.key_gen_time_formatted },
    { name: "Hashing", percent: perfMetrics.hashing_percent, color: "from-purple-500 to-purple-600", time: perfMetrics.hashing_time_formatted },
    { name: "Witness Generation", percent: perfMetrics.witness_percent, color: "from-green-500 to-green-600", time: perfMetrics.witness_time_formatted },
    { name: "Intersection", percent: perfMetrics.intersection_percent, color: "from-amber-500 to-amber-600", time: perfMetrics.intersection_time_formatted },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Performance Analytics</h2>
          <p className="text-slate-500">
            Real-time PSI operation performance and system metrics
          </p>
        </div>
        <Button onClick={fetchMetrics} variant="outline" disabled={loading}>
          <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          {loading ? 'Refreshing...' : 'Refresh'}
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Time</CardTitle>
            <Clock className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{perfMetrics.total_time_formatted}</div>
            <p className="text-xs text-slate-500">
              {perfMetrics.total_operations} operations
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Throughput</CardTitle>
            <Zap className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {perfMetrics.throughput_ops_per_sec.toFixed(2)}
            </div>
            <p className="text-xs text-slate-500">operations/sec</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Workers</CardTitle>
            <Cpu className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{perfMetrics.num_workers}</div>
            <p className="text-xs text-slate-500">parallel threads</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memory</CardTitle>
            <HardDrive className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{memMetrics.alloc_mb.toFixed(2)} MB</div>
            <p className="text-xs text-slate-500">{memMetrics.num_gc} GC cycles</p>
          </CardContent>
        </Card>
      </div>

      {/* Time Distribution */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Time Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {phases.map((phase) => (
                <div key={phase.name} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{phase.name}</span>
                    <span className="text-slate-600">
                      {phase.percent.toFixed(1)}% ({phase.time})
                    </span>
                  </div>
                  <div className="w-full bg-slate-100 rounded-full h-3 overflow-hidden">
                    <div
                      className={`h-full bg-gradient-to-r ${phase.color} transition-all duration-500`}
                      style={{ width: `${phase.percent}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Memory Usage Pie Chart (CSS-based) */}
        <Card>
          <CardHeader>
            <CardTitle>Memory Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 bg-gradient-to-br from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
                  <div className="text-xs text-blue-600 uppercase font-semibold mb-1">
                    Allocated
                  </div>
                  <div className="text-2xl font-bold text-blue-900">
                    {memMetrics.alloc_mb.toFixed(2)}
                  </div>
                  <div className="text-xs text-blue-500 mt-1">MB</div>
                </div>
                
                <div className="p-4 bg-gradient-to-br from-purple-50 to-pink-50 rounded-lg border border-purple-100">
                  <div className="text-xs text-purple-600 uppercase font-semibold mb-1">
                    Total Alloc
                  </div>
                  <div className="text-2xl font-bold text-purple-900">
                    {memMetrics.total_alloc_mb.toFixed(2)}
                  </div>
                  <div className="text-xs text-purple-500 mt-1">MB</div>
                </div>
                
                <div className="p-4 bg-gradient-to-br from-green-50 to-emerald-50 rounded-lg border border-green-100">
                  <div className="text-xs text-green-600 uppercase font-semibold mb-1">
                    System
                  </div>
                  <div className="text-2xl font-bold text-green-900">
                    {memMetrics.sys_mb.toFixed(2)}
                  </div>
                  <div className="text-xs text-green-500 mt-1">MB</div>
                </div>
                
                <div className="p-4 bg-gradient-to-br from-amber-50 to-orange-50 rounded-lg border border-amber-100">
                  <div className="text-xs text-amber-600 uppercase font-semibold mb-1">
                    Goroutines
                  </div>
                  <div className="text-2xl font-bold text-amber-900">
                    {memMetrics.goroutines}
                  </div>
                  <div className="text-xs text-amber-500 mt-1">active</div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
