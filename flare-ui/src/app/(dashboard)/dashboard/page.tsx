"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Users, ShieldCheck, Clock, Activity, FileUp } from "lucide-react";
import { apiClient, DashboardStats } from "@/lib/api-client";
import { formatDistanceToNow } from "date-fns";

import { isClientMode, isServerMode } from "@/lib/config";

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await apiClient.getDashboardStats();
        setStats(data);
      } catch (error) {
        console.error("Failed to fetch dashboard stats:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  if (loading) {
    return <div className="p-8">Loading dashboard...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
          <p className="text-slate-500">
            {isClientMode() 
              ? "Overview of screening operations and system status."
              : "Global sanctions authority overview and system status."}
          </p>
        </div>
        <div className="flex gap-3">
          {isClientMode() && (
            <>
              <Link href="/customers">
                <Button variant="outline">
                  <FileUp className="mr-2 h-4 w-4" />
                  Upload List
                </Button>
              </Link>
              <Link href="/screening">
                <Button>
                  <Activity className="mr-2 h-4 w-4" />
                  Run New Screening
                </Button>
              </Link>
            </>
          )}
          {isServerMode() && (
             <Link href="/sanctions">
                <Button>
                  <FileUp className="mr-2 h-4 w-4" />
                  Upload Sanctions
                </Button>
              </Link>
          )}
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {isClientMode() ? "Total Screenings" : "Active Sanction Lists"}
            </CardTitle>
            <Users className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {isClientMode() ? (stats?.totalScreenings || 0) : (stats?.activeLists || 0)}
            </div>
            <p className="text-xs text-slate-500">
              {isClientMode() ? "Lifetime screenings run" : "Global lists managed"}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {isClientMode() ? "Total Matches" : "Total Entities"}
            </CardTitle>
            <ShieldCheck className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.totalMatches || 0}</div>
            <p className="text-xs text-slate-500">
              {isClientMode() ? "Potential hits found" : "Sanctioned entities indexed"}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {isClientMode() ? "Last Screening" : "Last Update"}
            </CardTitle>
            <Clock className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats?.recentScreenings?.[0]?.finishedAt
                ? formatDistanceToNow(
                    new Date(stats.recentScreenings[0].finishedAt),
                    { addSuffix: true }
                  )
                : "Never"}
            </div>
            <p className="text-xs text-slate-500">
              {stats?.recentScreenings?.[0]?.name || "No recent activity"}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Status</CardTitle>
            <Activity className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {stats?.systemStatus || "Unknown"}
            </div>
            <p className="text-xs text-slate-500">
              {stats?.activeWorkers || 0} Workers Active
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity Section */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-8">
              {stats?.recentScreenings && stats.recentScreenings.length > 0 ? (
                stats.recentScreenings.map((screening) => (
                  <div key={screening.id} className="flex items-center">
                    <div className="h-9 w-9 rounded-full bg-slate-100 flex items-center justify-center border">
                      <Activity className="h-4 w-4 text-slate-500" />
                    </div>
                    <div className="ml-4 space-y-1">
                      <p className="text-sm font-medium leading-none">
                        {screening.name}
                      </p>
                      <p className="text-sm text-slate-500">
                        {new Date(screening.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    <div className="ml-auto font-medium flex items-center gap-4">
                      <span
                        className={
                          screening.matchCount > 0
                            ? "text-red-600"
                            : "text-slate-500"
                        }
                      >
                        {screening.matchCount} Matches
                      </span>
                      <span
                        className={`text-xs px-2 py-1 rounded-full ${
                          screening.status === "COMPLETED"
                            ? "bg-green-100 text-green-800"
                            : screening.status === "FAILED"
                            ? "bg-red-100 text-red-800"
                            : "bg-blue-100 text-blue-800"
                        }`}
                      >
                        {screening.status}
                      </span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-center text-slate-500 py-8">
                  No recent activity
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            {isClientMode() && (
              <>
                <Link href="/results">
                  <Button variant="secondary" className="w-full justify-start h-12">
                    <ShieldCheck className="mr-2 h-4 w-4" />
                    View All Matches
                  </Button>
                </Link>
                <Link href="/screening">
                  <Button variant="secondary" className="w-full justify-start h-12">
                    <Clock className="mr-2 h-4 w-4" />
                    Start New Scan
                  </Button>
                </Link>
              </>
            )}
            <Link href="/sanctions">
              <Button variant="secondary" className="w-full justify-start h-12">
                <FileUp className="mr-2 h-4 w-4" />
                {isServerMode() ? "Manage Sanctions Lists" : "View Sanctions Lists"}
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
