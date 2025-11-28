"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import {
  Play,
  Loader2,
  CheckCircle2,
  AlertCircle,
  FileText,
  Database,
  ShieldCheck,
  LucideIcon,
  LogOut,
} from "lucide-react";
import { cn } from "@/lib/utils";
import {
  apiClient,
  type ScreeningProgress,
  type CustomerList,
  type SanctionList,
  clearAccessToken,
} from "@/lib/api-client";

// Types for our data
type ScreeningStatus =
  | "idle"
  | "preparing"
  | "init_server"
  | "encrypting"
  | "intersecting"
  | "complete"
  | "failed";

export default function ScreeningPage() {
  const router = useRouter();
  const [status, setStatus] = useState<ScreeningStatus>("idle");
  const [progress, setProgress] = useState(0);
  const [logs, setLogs] = useState<string[]>([]);
  const [metrics, setMetrics] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [backendConnected, setBackendConnected] = useState(false);

  // Configuration State
  const [customerLists, setCustomerLists] = useState<CustomerList[]>([]);
  const [sanctionLists, setSanctionLists] = useState<SanctionList[]>([]);
  const [selectedCustomerListId, setSelectedCustomerListId] =
    useState<string>("");
  const [selectedSanctionListId, setSelectedSanctionListId] =
    useState<string>("");
  const [isUploading, setIsUploading] = useState(false);

  // Check backend connectivity and fetch lists on mount
  useEffect(() => {
    apiClient
      .healthCheck()
      .then(() => {
        setBackendConnected(true);
        addLog("Backend connected successfully");
        fetchLists();
      })
      .catch((err) => {
        setBackendConnected(false);
        addLog(`Backend connection failed: ${err.message}`);
        setError(
          "Backend server is not running. Please start the backend server."
        );
      });
  }, []);

  const handleLogout = () => {
    clearAccessToken();
    router.push("/login");
  };

  const fetchLists = async () => {
    try {
      const [cLists, sLists] = await Promise.all([
        apiClient.getCustomerLists(),
        apiClient.getSanctionLists(),
      ]);
      setCustomerLists(cLists);
      setSanctionLists(sLists);

      if (cLists.length > 0) setSelectedCustomerListId(cLists[0].id.toString());
      if (sLists.length > 0) setSelectedSanctionListId(sLists[0].id.toString());
    } catch (err) {
      console.error("Failed to fetch lists", err);
      // If fetch fails due to auth, the apiClient will redirect.
      // But we can also show a message if it's not auth related.
    }
  };

  const handleFileUpload = async (
    e: React.ChangeEvent<HTMLInputElement>,
    type: "customer" | "sanction"
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    addLog(`Uploading ${type} dataset: ${file.name}...`);

    try {
      if (type === "customer") {
        await apiClient.uploadCustomerList(file, file.name, "Uploaded via UI");
        addLog(`Customer dataset uploaded: ${file.name}`);
      } else {
        await apiClient.uploadSanctionList(
          file,
          file.name,
          "User Upload",
          "Uploaded via UI"
        );
        addLog(`Sanction dataset uploaded: ${file.name}`);
      }
      await fetchLists();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Unknown error";
      setError(`Upload failed: ${message}`);
      addLog(`Upload error: ${message}`);
    } finally {
      setIsUploading(false);
      // Reset input
      e.target.value = "";
    }
  };

  const addLog = (message: string) => {
    setLogs((prev) => [
      ...prev,
      `[${new Date().toLocaleTimeString()}] ${message}`,
    ]);
  };

  // Real API Integration
  const startScreening = async () => {
    if (!selectedCustomerListId || !selectedSanctionListId) return;
    if (!backendConnected) {
      setError("Backend server is not connected");
      return;
    }

    setError(null);
    setStatus("preparing");
    setProgress(5);

    const cList = customerLists.find(
      (l) => l.id.toString() === selectedCustomerListId
    );
    const sList = sanctionLists.find(
      (l) => l.id.toString() === selectedSanctionListId
    );

    addLog(`Job started: ${cList?.name} vs ${sList?.name}`);

    try {
      // Start the screening job
      const response = await apiClient.startScreening({
        name: `Screening ${cList?.name}`,
        customerListId: parseInt(selectedCustomerListId),
        sanctionListIds: [parseInt(selectedSanctionListId)],
      });

      addLog(`Job created: ${response.jobId}`);

      // Subscribe to real-time progress via SSE
      const unsubscribe = apiClient.subscribeToScreeningEvents(
        response.jobId,
        (progressData: ScreeningProgress) => {
          setStatus(progressData.stage); // Note: backend sends 'phase', frontend uses 'status' state but 'stage' in type? Wait, type mismatch.
          // Let's fix the type mismatch. Backend sends 'phase', 'percent', 'message', 'metrics'.
          // Frontend 'status' state expects specific strings.
          // We need to map backend 'phase' to frontend 'status'.
          
          // Actually, let's just use the phase from backend if it matches, or map it.
          // Backend phases: server_init, client_encrypt, intersection, persist, complete
          // Frontend statuses: idle, preparing, init_server, encrypting, intersecting, complete, failed
          
          let newStatus: ScreeningStatus = "idle";
          switch(progressData.phase) {
            case "server_init": newStatus = "init_server"; break;
            case "client_encrypt": newStatus = "encrypting"; break;
            case "intersection": newStatus = "intersecting"; break;
            case "persist": newStatus = "intersecting"; break; // Persist is part of intersection for UI
            case "complete": newStatus = "complete"; break;
            default: newStatus = "preparing";
          }
          setStatus(newStatus);
          setProgress(progressData.percent);
          addLog(progressData.message);
          if (progressData.metrics) {
            setMetrics(progressData.metrics);
          }
        },
        (err) => {
          setError(err.message);
          setStatus("failed");
          addLog(`Error: ${err.message}`);
        },
        () => {
          addLog("Screening complete");
        }
      );

      // Cleanup on unmount
      return () => {
        unsubscribe();
      };
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Unknown error";
      setError(message || "Failed to start screening");
      setStatus("failed");
      addLog(`Error: ${message}`);
    }
  };

  return (
    <div className="space-y-8 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Run Screening</h2>
          <p className="text-slate-500">
            Upload datasets and execute privacy-preserving intersection checks.
          </p>
        </div>
        {backendConnected ? (
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-sm text-green-600">
              <div className="h-2 w-2 bg-green-500 rounded-full animate-pulse"></div>
              Backend Connected
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleLogout}
              className="text-slate-500 hover:text-slate-900"
            >
              <LogOut className="h-4 w-4 mr-2" />
              Logout
            </Button>
          </div>
        ) : (
          <div className="flex items-center gap-2 text-sm text-red-600">
            <div className="h-2 w-2 bg-red-500 rounded-full"></div>
            Backend Disconnected
          </div>
        )}
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md">
          <p className="text-sm font-medium">{error}</p>
        </div>
      )}

      <div className="grid gap-8 md:grid-cols-3">
        {/* Configuration Panel */}
        <Card className="md:col-span-1 h-fit">
          <CardHeader>
            <CardTitle className="text-lg">Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Client Data Selection */}
            <div className="space-y-2">
              <label className="text-sm font-medium">
                Client Dataset (Customers)
              </label>
              <select
                className="w-full rounded-md border border-slate-300 p-2 text-sm"
                value={selectedCustomerListId}
                onChange={(e) => setSelectedCustomerListId(e.target.value)}
                disabled={status !== "idle" || isUploading}
              >
                <option value="">Select a list...</option>
                {customerLists.map((list) => (
                  <option key={list.id} value={list.id}>
                    {list.name} ({list.recordCount} recs)
                  </option>
                ))}
              </select>
              <div className="relative">
                <input
                  type="file"
                  id="customer-upload"
                  className="hidden"
                  accept=".csv"
                  onChange={(e) => handleFileUpload(e, "customer")}
                  disabled={status !== "idle" || isUploading}
                />
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full"
                  onClick={() =>
                    document.getElementById("customer-upload")?.click()
                  }
                  disabled={status !== "idle" || isUploading}
                >
                  {isUploading ? (
                    <Loader2 className="mr-2 h-3 w-3 animate-spin" />
                  ) : (
                    <FileText className="mr-2 h-3 w-3" />
                  )}
                  Upload New CSV
                </Button>
              </div>
            </div>

            {/* Server Data Selection */}
            <div className="space-y-2">
              <label className="text-sm font-medium">
                Server Dataset (Sanctions)
              </label>
              <select
                className="w-full rounded-md border border-slate-300 p-2 text-sm"
                value={selectedSanctionListId}
                onChange={(e) => setSelectedSanctionListId(e.target.value)}
                disabled={status !== "idle" || isUploading}
              >
                <option value="">Select a list...</option>
                {sanctionLists.map((list) => (
                  <option key={list.id} value={list.id}>
                    {list.name} ({list.recordCount} recs)
                  </option>
                ))}
              </select>

            </div>

            <Button
              className="w-full"
              onClick={startScreening}
              disabled={
                status !== "idle" ||
                !selectedCustomerListId ||
                !selectedSanctionListId ||
                isUploading
              }
            >
              {status === "idle" ? (
                <>
                  <Play className="mr-2 h-4 w-4" /> Start Screening
                </>
              ) : (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />{" "}
                  Processing...
                </>
              )}
            </Button>
          </CardContent>
        </Card>

        {/* Execution Status Panel */}
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="text-lg flex items-center justify-between">
              <span>Execution Status</span>
              {status === "complete" && (
                <Button
                  size="sm"
                  variant="outline"
                  className="text-green-600 border-green-200 bg-green-50"
                >
                  View Results
                </Button>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {/* Progress Bar */}
            <div className="mb-8 space-y-2">
              <div className="flex justify-between text-sm">
                <span>Overall Progress</span>
                <span className="font-mono">{progress}%</span>
              </div>
              <Progress value={progress} className="h-3" />
            </div>

            {/* Pipeline Stages */}
            <div className="grid grid-cols-4 gap-2 mb-8">
              <StageStep
                label="Prep"
                active={status === "preparing"}
                completed={[
                  "init_server",
                  "encrypting",
                  "intersecting",
                  "complete",
                ].includes(status)}
                icon={FileText}
              />
              <StageStep
                label="Init"
                active={status === "init_server"}
                completed={["encrypting", "intersecting", "complete"].includes(
                  status
                )}
                icon={Database}
              />
              <StageStep
                label="Encrypt"
                active={status === "encrypting"}
                completed={["intersecting", "complete"].includes(status)}
                icon={ShieldCheck}
              />
              <StageStep
                label="PSI Match"
                active={status === "intersecting"}
                completed={status === "complete"}
                icon={AlertCircle}
              />
            </div>

            {/* Terminal/Logs Area */}
            <div className="bg-slate-950 text-slate-50 rounded-md p-4 font-mono text-xs h-64 overflow-y-auto shadow-inner">
              {logs.length === 0 ? (
                <span className="text-slate-500">
                  {"// Ready for execution..."}
                </span>
              ) : (
                logs.map((log, i) => (
                  <div key={i} className="mb-1">
                    <span className="text-green-400">➜</span> {log}
                  </div>
                ))
              )}
              {status === "complete" && (
                <div className="mt-2 text-green-400 font-bold">DONE.</div>
              )}
            </div>

            {/* Metrics Grid */}
            {status !== "idle" && (
              <div className="grid grid-cols-3 gap-4 mt-6 pt-6 border-t">
                <div className="text-center">
                  <div className="text-xs text-slate-500 uppercase">
                    Throughput
                  </div>
                  <div className="text-lg font-bold">
                    {metrics["throughput"] || "0"}{" "}
                    <span className="text-xs font-normal text-slate-400">
                      rec/s
                    </span>
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-xs text-slate-500 uppercase">
                    Memory Usage
                  </div>
                  <div className="text-lg font-bold">
                    {metrics["memory"] || "0"}{" "}
                    <span className="text-xs font-normal text-slate-400">
                      MB
                    </span>
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-xs text-slate-500 uppercase">
                    CPU Load
                  </div>
                  <div className="text-lg font-bold">
                    {metrics["cpu"] || "0"}{" "}
                    <span className="text-xs font-normal text-slate-400">
                      %
                    </span>
                  </div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function StageStep({
  label,
  active,
  completed,
  icon: Icon,
}: {
  label: string;
  active: boolean;
  completed: boolean;
  icon: LucideIcon;
}) {
  return (
    <div
      className={cn(
        "flex flex-col items-center p-3 rounded-md border transition-colors",
        active
          ? "bg-blue-50 border-blue-200 text-blue-700"
          : completed
          ? "bg-green-50 border-green-200 text-green-700"
          : "bg-slate-50 border-slate-100 text-slate-400"
      )}
    >
      <div className="mb-2">
        {completed ? (
          <CheckCircle2 className="h-5 w-5" />
        ) : (
          <Icon className="h-5 w-5" />
        )}
      </div>
      <span className="text-xs font-medium">{label}</span>
    </div>
  );
}
