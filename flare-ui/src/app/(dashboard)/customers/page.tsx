"use client";

import { useEffect, useState, useRef } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Upload, Trash2, FileText, Users } from "lucide-react";
import { apiClient, CustomerList } from "@/lib/api-client";
import { formatDistanceToNow } from "date-fns";

export default function CustomersPage() {
  const [lists, setLists] = useState<CustomerList[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchLists = async () => {
    try {
      const data = await apiClient.getCustomerLists();
      setLists(data);
    } catch (error) {
      console.error("Failed to fetch customer lists:", error);
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
      formData.append(
        "description",
        `Uploaded on ${new Date().toLocaleDateString()}`
      );

      const uploadedFile = formData.get("file") as File;
      if (uploadedFile) {
        await apiClient.uploadCustomerList(uploadedFile, uploadedFile.name, "Uploaded via UI");
      }
      await fetchLists();

      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    } catch (error) {
      console.error("Failed to upload customer list:", error);
      alert("Failed to upload customer list. Please try again.");
    } finally {
      setUploading(false);
    }
  };

  if (loading) {
    return <div className="p-8">Loading customer lists...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Customer Lists</h2>
          <p className="text-slate-500">
            Manage customer data sources for screening.
          </p>
        </div>
        <div>
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
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Lists</CardTitle>
            <FileText className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{lists.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Records</CardTitle>
            <Users className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {lists.reduce((sum, list) => sum + (list.recordCount || 0), 0)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Latest Upload</CardTitle>
            <Upload className="h-4 w-4 text-slate-500" />
          </CardHeader>
          <CardContent>
            <div className="text-sm font-medium">
              {lists.length > 0
                ? formatDistanceToNow(new Date(lists[0].createdAt), {
                    addSuffix: true,
                  })
                : "No uploads yet"}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Lists Table */}
      <Card>
        <CardHeader>
          <CardTitle>Customer Lists</CardTitle>
        </CardHeader>
        <CardContent>
          {lists.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-center">
              <FileText className="h-12 w-12 text-slate-300 mb-4" />
              <p className="text-slate-400 mb-2">No customer lists yet</p>
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
                      Description
                    </th>
                    <th className="text-right p-4 font-medium text-slate-600">
                      Records
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
                      <td className="p-4 text-slate-600">{list.description}</td>
                      <td className="p-4 text-right text-slate-600">
                        {list.recordCount || 0}
                      </td>
                      <td className="p-4 text-slate-600">
                        {new Date(list.createdAt).toLocaleDateString()}
                      </td>
                      <td className="p-4 text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-red-600 hover:text-red-700 hover:bg-red-50"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
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
