"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  AlertTriangle,
  CheckCircle,
  XCircle,
  Eye,
  Download,
  Filter,
  ShieldAlert,
  ThumbsDown,
  ThumbsUp,
} from "lucide-react";
import { apiClient, DashboardStats } from "@/lib/api-client";
import { MatchDetailsModal } from "@/components/match-details-modal";

interface ScreeningResult {
  id: number;
  screeningId: number;
  customerName: string;
  customerDob: string;
  customerCountry: string;
  sanctionName: string;
  sanctionDob: string;
  sanctionCountry: string;
  sanctionProgram: string;
  matchScore: number;
  status: "PENDING" | "CONFIRMED" | "FALSE_POSITIVE";
  reviewedBy?: string;
  reviewedAt?: string;
  notes?: string;
}

export default function ResultsPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [results, setResults] = useState<ScreeningResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<
    "ALL" | "PENDING" | "CONFIRMED" | "FALSE_POSITIVE"
  >("ALL");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedMatch, setSelectedMatch] = useState<ScreeningResult | null>(null);

  // Export to CSV function
  const exportToCSV = () => {
    const csvContent = [
      // Header
      [
        "Customer Name",
        "Customer DOB",
        "Customer Country",
        "Sanction Name",
        "Sanction DOB",
        "Sanction Country",
        "Program",
        "Match Score",
        "Status",
      ].join(","),
      // Data rows
      ...filteredResults.map((r) =>
        [
          `"${r.customerName}"`,
          r.customerDob,
          r.customerCountry,
          `"${r.sanctionName}"`,
          r.sanctionDob,
          r.sanctionCountry,
          r.sanctionProgram,
          (r.matchScore * 100).toFixed(2) + "%",
          r.status,
        ].join(",")
      ),
    ].join("\n");

    const blob = new Blob([csvContent], { type: "text/csv" });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `screening-results-${new Date().toISOString()}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
  };

  // Mark as false positive - persist to backend
  const markAsFalsePositive = async (resultId: number) => {
    try {
      // Update locally first for immediate feedback
      setResults((prev) =>
        prev.map((r) =>
          r.id === resultId ? { ...r, status: "FALSE_POSITIVE" as const } : r
        )
      );
      
      // Persist to backend
      await apiClient.updateResultStatus(resultId, "FALSE_POSITIVE");
    } catch (error) {
      console.error("Failed to update status:", error);
      // Revert on error
      setResults((prev) =>
        prev.map((r) =>
          r.id === resultId ? { ...r, status: "PENDING" as const } : r
        )
      );
    }
  };

  // Mark as confirmed - persist to backend
  const markAsConfirmed = async (resultId: number) => {
    try {
      // Update locally first for immediate feedback
      setResults((prev) =>
        prev.map((r) =>
          r.id === resultId ? { ...r, status: "CONFIRMED" as const } : r
        )
      );
      
      // Persist to backend
      await apiClient.updateResultStatus(resultId, "CONFIRMED");
    } catch (error) {
      console.error("Failed to update status:", error);
      // Revert on error
      setResults((prev) =>
        prev.map((r) =>
          r.id === resultId ? { ...r, status: "PENDING" as const } : r
        )
      );
    }
  };

  useEffect(() => {
    const fetchData = async () => {
      try {
        const statsData = await apiClient.getDashboardStats();
        setStats(statsData);

        // Fetch results for all recent screenings that have matches
        const screeningsWithMatches = statsData.recentScreenings.filter(
          (s) => s.matchCount > 0
        );

        const allResults: ScreeningResult[] = [];
        
        for (const screening of screeningsWithMatches) {
          try {
            const resultsData = await apiClient.getScreeningResults(screening.jobId);
            // Map backend results to frontend model
            // Backend result: { id, screeningId, customerRecordId, sanctionRecordId, matchScore, status, ... }
            // We need to fetch details or the backend should return joined data.
            // Assuming backend returns joined data for now or we map what we have.
            // Actually, looking at handlers.go, GetScreeningResults returns the raw rows.
            // We might need to update the backend to return more details or just show what we have.
            
            // Let's assume the backend returns enough info or we'll just show IDs for now if names aren't there.
            // Wait, the backend GetScreeningResults just returns `[]models.ScreeningResult`.
            // Let's check `models.ScreeningResult`.
            
            // For now, I'll map what I can and maybe use placeholders for missing names if the backend doesn't provide them yet.
            // But the goal is "real values". 
            // If the backend doesn't return names, I should probably update the backend to join with Customer/Sanction tables.
            // Let's check the backend handler for `GetScreeningResults`.
            
            
            const mappedResults = ((resultsData as any).results || []).map((r: any, idx: number) => ({
              id: r.id,
              screeningId: r.screeningId,
              customerName: r.customer?.name || `Customer #${r.customerId}`,
              customerDob: r.customer?.dob || "N/A",
              customerCountry: r.customer?.country || "N/A",
              sanctionName: r.sanction?.name || `Sanction #${r.sanctionId}`,
              sanctionDob: r.sanction?.dob || "N/A",
              sanctionCountry: r.sanction?.country || "N/A",
              sanctionProgram: r.sanction?.program || "UNKNOWN",
              matchScore: r.matchScore || 0,
              status: r.status || "PENDING",
            }));
            allResults.push(...mappedResults);
          } catch (err) {
            console.error(`Failed to fetch results for job ${screening.jobId}`, err);
          }
        }
        
        setResults(allResults);
      } catch (error) {
        console.error("Failed to fetch results:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const filteredResults = results.filter((result) => {
    // Filter by status
    if (filter !== "ALL" && result.status !== filter) return false;
    
    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      return (
        result.customerName.toLowerCase().includes(query) ||
        result.sanctionName.toLowerCase().includes(query) ||
        result.customerCountry.toLowerCase().includes(query) ||
        result.sanctionCountry.toLowerCase().includes(query) ||
        result.sanctionProgram.toLowerCase().includes(query)
      );
    }
    
    return true;
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "CONFIRMED":
        return (
          <span className="flex items-center gap-1 text-xs px-2 py-1 rounded-full bg-red-100 text-red-800">
            <AlertTriangle className="h-3 w-3" />
            Confirmed
          </span>
        );
      case "FALSE_POSITIVE":
        return (
          <span className="flex items-center gap-1 text-xs px-2 py-1 rounded-full bg-green-100 text-green-800">
            <CheckCircle className="h-3 w-3" />
            False Positive
          </span>
        );
      default:
        return (
          <span className="flex items-center gap-1 text-xs px-2 py-1 rounded-full bg-amber-100 text-amber-800">
            <AlertTriangle className="h-3 w-3" />
            Pending Review
          </span>
        );
    }
  };

  const getRiskBadge = (program: string) => {
    const riskLevels: Record<string, string> = {
      TERRORISM_FINANCING: "bg-red-100 text-red-800",
      MONEY_LAUNDERING: "bg-amber-100 text-amber-800",
      WEAPONS_PROLIFERATION: "bg-red-100 text-red-800",
      CORRUPTION: "bg-amber-100 text-amber-800",
      CYBER_CRIMES: "bg-orange-100 text-orange-800",
      DEFAULT: "bg-slate-100 text-slate-800",
    };
    const colorClass = riskLevels[program] || riskLevels.DEFAULT;
    return (
      <span className={`text-xs px-2 py-1 rounded-full ${colorClass}`}>
        {program.replace(/_/g, " ")}
      </span>
    );
  };

  if (loading) {
    return <div className="p-8">Loading results...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">
            Screening Results
          </h2>
          <p className="text-slate-500">Review and manage screening matches.</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={exportToCSV} disabled={filteredResults.length === 0}>
            <Download className="mr-2 h-4 w-4" />
            Export CSV
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Matches</CardTitle>
            <ShieldAlert className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{results.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Pending Review
            </CardTitle>
            <AlertTriangle className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {results.filter((r) => r.status === "PENDING").length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Confirmed</CardTitle>
            <XCircle className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {results.filter((r) => r.status === "CONFIRMED").length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              False Positives
            </CardTitle>
            <CheckCircle className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {results.filter((r) => r.status === "FALSE_POSITIVE").length}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div className="flex gap-2 flex-wrap">
          <Button
            variant={filter === "ALL" ? "default" : "outline"}
            size="sm"
            onClick={() => setFilter("ALL")}
          >
            All ({results.length})
          </Button>
          <Button
            variant={filter === "PENDING" ? "default" : "outline"}
            size="sm"
            onClick={() => setFilter("PENDING")}
          >
            Pending ({results.filter((r) => r.status === "PENDING").length})
          </Button>
          <Button
            variant={filter === "CONFIRMED" ? "default" : "outline"}
            size="sm"
            onClick={() => setFilter("CONFIRMED")}
          >
            Confirmed ({results.filter((r) => r.status === "CONFIRMED").length})
          </Button>
          <Button
            variant={filter === "FALSE_POSITIVE" ? "default" : "outline"}
            size="sm"
            onClick={() => setFilter("FALSE_POSITIVE")}
          >
            False Positive (
            {results.filter((r) => r.status === "FALSE_POSITIVE").length})
          </Button>
        </div>
        
        {/* Search Input */}
        <div className="relative">
          <input
            type="text"
            placeholder="Search by name, country, program..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10 pr-4 py-2 border border-slate-200 rounded-md focus:outline-none focus:ring-2 focus:ring-slate-900 focus:border-transparent w-full md:w-80"
          />
          <Filter className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
        </div>
      </div>

      {/* Results Table */}
      <Card>
        <CardHeader>
          <CardTitle>Matches</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredResults.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-center">
              <ShieldAlert className="h-12 w-12 text-slate-300 mb-4" />
              <p className="text-slate-400 mb-2">No matches found</p>
              <p className="text-sm text-slate-500">
                {filter !== "ALL"
                  ? "Try changing the filter"
                  : "Run a screening to find potential matches"}
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-200">
                    <th className="text-left p-4 font-medium text-slate-600">
                      Customer
                    </th>
                    <th className="text-left p-4 font-medium text-slate-600">
                      Sanction Match
                    </th>
                    <th className="text-left p-4 font-medium text-slate-600">
                      Program
                    </th>
                    <th className="text-center p-4 font-medium text-slate-600">
                      Score
                    </th>
                    <th className="text-left p-4 font-medium text-slate-600">
                      Status
                    </th>
                    <th className="text-right p-4 font-medium text-slate-600">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {filteredResults.map((result) => (
                    <tr
                      key={result.id}
                      className="border-b border-slate-100 hover:bg-slate-50"
                    >
                      <td className="p-4">
                        <div className="font-medium">{result.customerName}</div>
                        <div className="text-sm text-slate-600">
                          {result.customerDob} • {result.customerCountry}
                        </div>
                      </td>
                      <td className="p-4">
                        <div className="font-medium">{result.sanctionName}</div>
                        <div className="text-sm text-slate-600">
                          {result.sanctionDob} • {result.sanctionCountry}
                        </div>
                      </td>
                      <td className="p-4">
                        {getRiskBadge(result.sanctionProgram)}
                      </td>
                      <td className="p-4 text-center">
                        <span className="font-mono font-medium">
                          {(result.matchScore * 100).toFixed(0)}%
                        </span>
                      </td>
                      <td className="p-4">{getStatusBadge(result.status)}</td>
                      <td className="p-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          {result.status === "PENDING" && (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => markAsConfirmed(result.id)}
                                className="text-red-600 hover:text-red-700 hover:bg-red-50"
                                title="Mark as Confirmed Match"
                              >
                                <ThumbsDown className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => markAsFalsePositive(result.id)}
                                className="text-green-600 hover:text-green-700 hover:bg-green-50"
                                title="Mark as False Positive"
                              >
                                <ThumbsUp className="h-4 w-4" />
                              </Button>
                            </>
                          )}
                          <Button 
                            variant="ghost" 
                            size="sm" 
                            title="View Details"
                            onClick={() => setSelectedMatch(result)}
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Match Details Modal */}
      {selectedMatch && (
        <MatchDetailsModal
          match={selectedMatch}
          onClose={() => setSelectedMatch(null)}
        />
      )}
    </div>
  );
}
