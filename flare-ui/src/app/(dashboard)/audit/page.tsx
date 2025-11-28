"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Activity,
  Upload,
  ShieldCheck,
  UserCheck,
  Clock,
  Filter,
  Download,
} from "lucide-react";
import { apiClient, DashboardStats } from "@/lib/api-client";
import { formatDistanceToNow } from "date-fns";

interface AuditEntry {
  id: number;
  timestamp: string;
  action: string;
  entity: string;
  entityId: string;
  user: string;
  details: string;
  severity: "INFO" | "WARNING" | "CRITICAL";
}

export default function AuditPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [auditLog, setAuditLog] = useState<AuditEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<"ALL" | "INFO" | "WARNING" | "CRITICAL">(
    "ALL"
  );

  useEffect(() => {
    const fetchData = async () => {
      try {
        const statsData = await apiClient.getDashboardStats();
        setStats(statsData);

        // Generate mock audit entries from recent screenings
        const mockEntries: AuditEntry[] = [];
        statsData.recentScreenings?.forEach((screening, idx) => {
          mockEntries.push({
            id: idx * 3 + 1,
            timestamp: screening.createdAt,
            action: "SCREENING_STARTED",
            entity: "Screening",
            entityId: screening.jobId,
            user: "Admin User",
            details: `Started screening: ${screening.name}`,
            severity: "INFO",
          });

          if (screening.status === "COMPLETED") {
            mockEntries.push({
              id: idx * 3 + 2,
              timestamp: screening.finishedAt,
              action: "SCREENING_COMPLETED",
              entity: "Screening",
              entityId: screening.jobId,
              user: "System",
              details: `Completed with ${screening.matchCount} matches`,
              severity: screening.matchCount > 0 ? "WARNING" : "INFO",
            });
          }

          if (screening.matchCount > 0) {
            mockEntries.push({
              id: idx * 3 + 3,
              timestamp: screening.finishedAt,
              action: "MATCH_DETECTED",
              entity: "Result",
              entityId: `match-${idx}`,
              user: "System",
              details: `Detected ${screening.matchCount} potential match(es)`,
              severity: "CRITICAL",
            });
          }
        });

        setAuditLog(
          mockEntries.sort(
            (a, b) =>
              new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
          )
        );
      } catch (error) {
        console.error("Failed to fetch audit data:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const filteredLog = auditLog.filter((entry) => {
    if (filter === "ALL") return true;
    return entry.severity === filter;
  });

  const getActionIcon = (action: string) => {
    switch (action) {
      case "SCREENING_STARTED":
      case "SCREENING_COMPLETED":
        return <ShieldCheck className="h-4 w-4" />;
      case "MATCH_DETECTED":
        return <Activity className="h-4 w-4" />;
      case "LIST_UPLOADED":
        return <Upload className="h-4 w-4" />;
      case "USER_LOGIN":
        return <UserCheck className="h-4 w-4" />;
      default:
        return <Activity className="h-4 w-4" />;
    }
  };

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case "CRITICAL":
        return "bg-red-100 text-red-800";
      case "WARNING":
        return "bg-amber-100 text-amber-800";
      default:
        return "bg-blue-100 text-blue-800";
    }
  };

  if (loading) {
    return <div className="p-8">Loading audit log...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Audit Log</h2>
          <p className="text-slate-500">
            Track all system activities and user actions.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline">
            <Download className="mr-2 h-4 w-4" />
            Export Log
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Events</CardTitle>
            <Activity className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{auditLog.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Critical</CardTitle>
            <Activity className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {auditLog.filter((e) => e.severity === "CRITICAL").length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Warnings</CardTitle>
            <Activity className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {auditLog.filter((e) => e.severity === "WARNING").length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Last Event</CardTitle>
            <Clock className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-sm font-medium">
              {auditLog.length > 0
                ? formatDistanceToNow(new Date(auditLog[0].timestamp), {
                    addSuffix: true,
                  })
                : "No events"}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <div className="flex gap-2">
        <Button
          variant={filter === "ALL" ? "default" : "outline"}
          size="sm"
          onClick={() => setFilter("ALL")}
        >
          All ({auditLog.length})
        </Button>
        <Button
          variant={filter === "INFO" ? "default" : "outline"}
          size="sm"
          onClick={() => setFilter("INFO")}
        >
          Info ({auditLog.filter((e) => e.severity === "INFO").length})
        </Button>
        <Button
          variant={filter === "WARNING" ? "default" : "outline"}
          size="sm"
          onClick={() => setFilter("WARNING")}
        >
          Warning ({auditLog.filter((e) => e.severity === "WARNING").length})
        </Button>
        <Button
          variant={filter === "CRITICAL" ? "default" : "outline"}
          size="sm"
          onClick={() => setFilter("CRITICAL")}
        >
          Critical ({auditLog.filter((e) => e.severity === "CRITICAL").length})
        </Button>
      </div>

      {/* Audit Log Table */}
      <Card>
        <CardHeader>
          <CardTitle>Activity Log</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredLog.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-center">
              <Activity className="h-12 w-12 text-slate-300 mb-4" />
              <p className="text-slate-400 mb-2">No events found</p>
              <p className="text-sm text-slate-500">
                {filter !== "ALL"
                  ? "Try changing the filter"
                  : "System activity will appear here"}
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {filteredLog.map((entry) => (
                <div
                  key={entry.id}
                  className="flex items-start gap-4 p-4 rounded-lg border border-slate-100 hover:bg-slate-50"
                >
                  <div
                    className={`mt-1 p-2 rounded-lg ${getSeverityBadge(
                      entry.severity
                    )}`}
                  >
                    {getActionIcon(entry.action)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-medium text-slate-900">
                        {entry.action.replace(/_/g, " ")}
                      </span>
                      <span
                        className={`text-xs px-2 py-0.5 rounded-full ${getSeverityBadge(
                          entry.severity
                        )}`}
                      >
                        {entry.severity}
                      </span>
                    </div>
                    <p className="text-sm text-slate-600 mb-1">
                      {entry.details}
                    </p>
                    <div className="flex items-center gap-4 text-xs text-slate-500">
                      <span className="flex items-center gap-1">
                        <UserCheck className="h-3 w-3" />
                        {entry.user}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {formatDistanceToNow(new Date(entry.timestamp), {
                          addSuffix: true,
                        })}
                      </span>
                      <span className="text-slate-400">
                        {entry.entity} #{entry.entityId}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
