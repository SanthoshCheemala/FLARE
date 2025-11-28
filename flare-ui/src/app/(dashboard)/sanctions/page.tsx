"use client";

import { useEffect, useState, useRef } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Upload, Trash2, FileText, ShieldAlert, Database } from "lucide-react";
import { apiClient, SanctionList } from "@/lib/api-client";
import { formatDistanceToNow } from "date-fns";

import { isServerMode } from "@/lib/config";

export default function SanctionsPage() {
  const [lists, setLists] = useState<SanctionList[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchLists = async () => {
    try {
      const data = await apiClient.getSanctionLists();
      setLists(data);
    } catch (error) {
      console.error("Failed to fetch sanction lists:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLists();
  }, []);

  const handleUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setUploading(true);
    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("name", file.name.replace(".csv", ""));
      formData.append("source", "Manual Upload");
      formData.append(
        "description",
        `Uploaded on ${new Date().toLocaleDateString()}`
      );

      const uploadedFile = formData.get("file") as File;
      if (uploadedFile) {
        await apiClient.uploadSanctionList(
          uploadedFile, 
          uploadedFile.name, 
          "User Upload", 
          "Uploaded via UI"
        );
      }
      await fetchLists();

      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    } catch (error) {
      console.error("Failed to upload sanction list:", error);
      alert("Failed to upload sanction list. Please try again.");
    } finally {
      setUploading(false);
    }
  };

  const getRiskBadgeColor = (source: string) => {
    const lowerSource = source.toLowerCase();
    if (lowerSource.includes("ofac") || lowerSource.includes("un")) {
      return "bg-red-100 text-red-800";
    }
    if (lowerSource.includes("eu")) {
      return "bg-amber-100 text-amber-800";
    }
    return "bg-slate-100 text-slate-800";
  };

  if (loading) {
    return <div className="p-8">Loading sanction lists...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Sanctions Lists</h1>
          <p className="text-muted-foreground">
            {isServerMode() 
              ? "Manage global sanctions lists for the Authority." 
              : "View available sanctions lists from the Sanctions Authority."}
          </p>
        </div>
        <div>
          {isServerMode() && (
            <>
              <input
                ref={fileInputRef}
                type="file"
                accept=".csv"
                onChange={handleUpload}
                className="hidden"
              />
              <Button
                onClick={() => fileInputRef.current?.click()}
                disabled={uploading}
              >
                <Upload className="mr-2 h-4 w-4" />
                {uploading ? "Uploading..." : "Upload CSV"}
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Lists</CardTitle>
            <Database className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{lists.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Sanctioned Entities
            </CardTitle>
            <ShieldAlert className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {lists.reduce((sum, list) => sum + (list.recordCount || 0), 0)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Last Updated</CardTitle>
            <Upload className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-sm font-medium">
              {lists.length > 0 ? (() => {
                // Find the most recent date
                const sortedLists = [...lists].sort((a, b) => 
                  new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
                );
                const latestDateStr = sortedLists[0]?.createdAt;
                
                if (!latestDateStr) return "No updates yet";

                const date = new Date(latestDateStr);
                if (isNaN(date.getTime())) return "Invalid date";

                return formatDistanceToNow(date, { addSuffix: true });
              })() : "No updates yet"}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Lists Table */}
      <Card>
        <CardHeader>
          <CardTitle>Sanctions Databases</CardTitle>
        </CardHeader>
        <CardContent>
          {lists.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-center">
              <ShieldAlert className="h-12 w-12 text-slate-300 mb-4" />
              <p className="text-slate-400 mb-2">No sanction lists yet</p>
              <p className="text-sm text-slate-500">
                Upload a CSV file to get started
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-200">
                    <th className="text-left p-4 font-medium text-slate-600">
                      Name
                    </th>
                    <th className="text-left p-4 font-medium text-slate-600">
                      Source
                    </th>
                    <th className="text-left p-4 font-medium text-slate-600">
                      Description
                    </th>
                    <th className="text-right p-4 font-medium text-slate-600">
                      Entities
                    </th>
                    <th className="text-left p-4 font-medium text-slate-600">
                      Created
                    </th>
                    <th className="text-right p-4 font-medium text-slate-600">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {lists.map((list) => (
                    <tr
                      key={list.id}
                      className="border-b border-slate-100 hover:bg-slate-50"
                    >
                      <td className="p-4 font-medium">{list.name}</td>
                      <td className="p-4">
                        <span
                          className={`text-xs px-2 py-1 rounded-full ${getRiskBadgeColor(
                            list.source
                          )}`}
                        >
                          {list.source}
                        </span>
                      </td>
                      <td className="p-4 text-slate-600">{list.description}</td>
                      <td className="p-4 text-right text-slate-600">
                        {list.recordCount || 0}
                      </td>
                      <td className="p-4 text-slate-600">
                        {new Date(list.createdAt).toLocaleDateString()}
                      </td>
                      <td className="p-4 text-right">
                        {isServerMode() && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-red-600 hover:text-red-700 hover:bg-red-50"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
