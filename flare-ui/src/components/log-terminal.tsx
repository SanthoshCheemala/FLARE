"use client";

import { useEffect, useRef, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";

export function LogTerminal() {
  const [logs, setLogs] = useState<string[]>([]);
  const [status, setStatus] = useState<"connected" | "disconnected" | "connecting">("connecting");
  const scrollRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const connect = () => {
      // Use port 8080 for backend
      const wsUrl = "ws://localhost:8080/ws/logs";
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setStatus("connected");
        setLogs((prev) => [...prev, "Connected to log stream..."]);
      };

      ws.onmessage = (event) => {
        setLogs((prev) => {
          const newLogs = [...prev, event.data];
          // Keep last 1000 lines
          if (newLogs.length > 1000) {
            return newLogs.slice(newLogs.length - 1000);
          }
          return newLogs;
        });
      };

      ws.onclose = () => {
        setStatus("disconnected");
        // Reconnect after 3 seconds
        setTimeout(connect, 3000);
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        ws.close();
      };
    };

    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs]);

  return (
    <Card className="w-full mt-6 bg-black border-slate-800">
      <CardHeader className="flex flex-row items-center justify-between py-3">
        <CardTitle className="text-sm font-mono text-slate-300">Backend Logs</CardTitle>
        <Badge 
          variant="outline" 
          className={
            status === "connected" ? "text-green-500 border-green-900 bg-green-950/30" : 
            status === "connecting" ? "text-yellow-500 border-yellow-900 bg-yellow-950/30" : 
            "text-red-500 border-red-900 bg-red-950/30"
          }
        >
          {status === "connected" ? "● Live" : status === "connecting" ? "○ Connecting" : "○ Disconnected"}
        </Badge>
      </CardHeader>
      <CardContent className="p-0">
        <div 
          ref={scrollRef}
          className="h-[300px] overflow-y-auto p-4 font-mono text-xs text-slate-400 space-y-1"
        >
          {logs.map((log, i) => (
            <div key={i} className="break-all whitespace-pre-wrap border-b border-slate-900/50 pb-0.5">
              {log}
            </div>
          ))}
          {logs.length === 0 && (
            <div className="text-slate-600 italic">Waiting for logs...</div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
