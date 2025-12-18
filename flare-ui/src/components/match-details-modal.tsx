"use client";

import { Button } from "@/components/ui/button";
import { X, User, Shield, AlertCircle } from "lucide-react";

interface MatchDetailsProps {
  match: {
    id: number;
    customerName: string;
    customerDob: string;
    customerCountry: string;
    sanctionName: string;
    sanctionDob: string;
    sanctionCountry: string;
    sanctionProgram: string;
    matchScore: number;
    status: string;
  };
  onClose: () => void;
}

export function MatchDetailsModal({ match, onClose }: MatchDetailsProps) {
  const exportReport = () => {
    // Generate detailed report content
    const reportContent = `
SANCTIONS SCREENING MATCH REPORT
Generated: ${new Date().toLocaleString()}
=====================================

MATCH OVERVIEW
--------------
Match ID: ${match.id}
Match Confidence: ${(match.matchScore * 100).toFixed(2)}%
Status: ${match.status.replace('_', ' ')}
Risk Level: ${match.sanctionProgram.includes("TERRORISM") || match.sanctionProgram.includes("WEAPONS") ? "HIGH" : "MEDIUM"}

CUSTOMER INFORMATION
--------------------
Full Name: ${match.customerName}
Date of Birth: ${match.customerDob}
Country: ${match.customerCountry}

SANCTION INFORMATION
--------------------
Full Name: ${match.sanctionName}
Date of Birth: ${match.sanctionDob}
Country: ${match.sanctionCountry}
Sanction Program: ${match.sanctionProgram.replace(/_/g, ' ')}

MATCH ANALYSIS
--------------
Name Match: ${match.customerName.toLowerCase() === match.sanctionName.toLowerCase() ? "EXACT" : "PARTIAL"}
DOB Match: ${match.customerDob === match.sanctionDob ? "YES" : "NO"}
Country Match: ${match.customerCountry === match.sanctionCountry ? "YES" : "NO"}

RECOMMENDATION
--------------
${match.status === "PENDING" ? "This match requires manual review and decision." : 
  match.status === "CONFIRMED" ? "This match has been CONFIRMED as a true positive." :
  "This match has been marked as a FALSE POSITIVE."}

=====================================
FLARE - LE-PSI Sanctions Screening System
Confidential - For Authorized Use Only
    `.trim();

    // Create and download file
    const blob = new Blob([reportContent], { type: 'text/plain' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `match-report-${match.id}-${new Date().toISOString().split('T')[0]}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b bg-gradient-to-r from-slate-50 to-slate-100">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-amber-100 flex items-center justify-center">
              <AlertCircle className="h-5 w-5 text-amber-600" />
            </div>
            <div>
              <h2 className="text-xl font-bold">Match Details</h2>
              <p className="text-sm text-slate-500">Review screening result</p>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-180px)]">
          {/* Match Score Banner */}
          <div className="mb-6 p-4 bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-200 rounded-lg">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm text-amber-700 font-medium mb-1">Match Confidence</div>
                <div className="text-3xl font-bold text-amber-900">
                  {(match.matchScore * 100).toFixed(0)}%
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm text-amber-700 font-medium mb-1">Status</div>
                <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${
                  match.status === "CONFIRMED"
                    ? "bg-red-100 text-red-800"
                    : match.status === "FALSE_POSITIVE"
                    ? "bg-green-100 text-green-800"
                    : "bg-amber-100 text-amber-800"
                }`}>
                  {match.status.replace('_', ' ')}
                </span>
              </div>
            </div>
          </div>

          {/* Side-by-side Comparison */}
          <div className="grid md:grid-cols-2 gap-6">
            {/* Customer Details */}
            <div className="border rounded-lg p-5 bg-blue-50/30">
              <div className="flex items-center gap-2 mb-4 pb-3 border-b">
                <User className="h-5 w-5 text-blue-600" />
                <h3 className="font-semibold text-blue-900">Customer Information</h3>
              </div>
              <div className="space-y-3">
                <div>
                  <div className="text-xs text-blue-600 font-medium mb-1">Full Name</div>
                  <div className="text-sm font-semibold text-blue-900">{match.customerName}</div>
                </div>
                <div>
                  <div className="text-xs text-blue-600 font-medium mb-1">Date of Birth</div>
                  <div className="text-sm text-blue-800">{match.customerDob}</div>
                </div>
                <div>
                  <div className="text-xs text-blue-600 font-medium mb-1">Country</div>
                  <div className="text-sm text-blue-800">{match.customerCountry}</div>
                </div>
              </div>
            </div>

            {/* Sanction Details */}
            <div className="border rounded-lg p-5 bg-red-50/30">
              <div className="flex items-center gap-2 mb-4 pb-3 border-b">
                <Shield className="h-5 w-5 text-red-600" />
                <h3 className="font-semibold text-red-900">Sanction Information</h3>
              </div>
              <div className="space-y-3">
                <div>
                  <div className="text-xs text-red-600 font-medium mb-1">Full Name</div>
                  <div className="text-sm font-semibold text-red-900">{match.sanctionName}</div>
                </div>
                <div>
                  <div className="text-xs text-red-600 font-medium mb-1">Date of Birth</div>
                  <div className="text-sm text-red-800">{match.sanctionDob}</div>
                </div>
                <div>
                  <div className="text-xs text-red-600 font-medium mb-1">Country</div>
                  <div className="text-sm text-red-800">{match.sanctionCountry}</div>
                </div>
                <div>
                  <div className="text-xs text-red-600 font-medium mb-1">Sanction Program</div>
                  <div className="text-sm text-red-800 font-medium">
                    {match.sanctionProgram.replace(/_/g, ' ')}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Match Analysis */}
          <div className="mt-6 p-4 bg-slate-50 border rounded-lg">
            <h4 className="font-semibold text-slate-900 mb-3">Match Analysis</h4>
            <div className="space-y-2 text-sm text-slate-700">
              <div className="flex items-center justify-between">
                <span>Name Similarity:</span>
                <span className="font-medium">
                  {match.customerName.toLowerCase() === match.sanctionName.toLowerCase() ? "Exact" : "Partial"}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span>DOB Match:</span>
                <span className="font-medium">
                  {match.customerDob === match.sanctionDob ? "Yes" : "No"}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span>Country Match:</span>
                <span className="font-medium">
                  {match.customerCountry === match.sanctionCountry ? "Yes" : "No"}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span>Risk Level:</span>
                <span className={`font-medium ${
                  match.sanctionProgram.includes("TERRORISM") || match.sanctionProgram.includes("WEAPONS")
                    ? "text-red-600"
                    : "text-amber-600"
                }`}>
                  {match.sanctionProgram.includes("TERRORISM") || match.sanctionProgram.includes("WEAPONS")
                    ? "High"
                    : "Medium"}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-6 border-t bg-slate-50 flex justify-end gap-3">
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
          <Button 
            className="bg-blue-600 hover:bg-blue-700"
            onClick={exportReport}
          >
            Export Report
          </Button>
        </div>
      </div>
    </div>
  );
}
