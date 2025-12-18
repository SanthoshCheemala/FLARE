"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Settings, Zap, Info, Cpu, HardDrive } from "lucide-react";
import { useState } from "react";

interface AdvancedConfigProps {
  datasetSize: number;
  onConfigChange?: (config: ScreeningConfig) => void;
}

export interface ScreeningConfig {
  numWorkers: number;
  memoryLimit: number; // in MB
  autoOptimize: boolean;
  batchSize: number;
}

export function AdvancedConfig({ datasetSize, onConfigChange }: AdvancedConfigProps) {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [config, setConfig] = useState<ScreeningConfig>({
    numWorkers: calculateOptimalWorkers(datasetSize),
    memoryLimit: 8192, // 8GB default
    autoOptimize: true,
    batchSize: 1000,
  });

  // Calculate optimal workers based on dataset size
  function calculateOptimalWorkers(size: number): number {
    if (size < 1000) return 8;
    if (size < 5000) return 16;
    if (size < 10000) return 24;
    if (size < 50000) return 32;
    return 48; // Max for large datasets
  }

  const handleConfigUpdate = (field: keyof ScreeningConfig, value: number | boolean) => {
    const newConfig = { ...config, [field]: value };
    setConfig(newConfig);
    onConfigChange?.(newConfig);
  };

  const estimatedRAM = (datasetSize / 30) * config.numWorkers; // ~35MB per record with workers
  const estimatedTime = datasetSize / (config.numWorkers * 100); // Rough estimate

  return (
    <Card className="border-blue-200 bg-blue-50/30">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-medium flex items-center gap-2">
            <Settings className="h-4 w-4" />
            Advanced Configuration
          </CardTitle>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowAdvanced(!showAdvanced)}
          >
            {showAdvanced ? "Hide" : "Show"}
          </Button>
        </div>
      </CardHeader>
      
      {showAdvanced && (
        <CardContent className="space-y-6">
          {/* Auto-Optimize Toggle */}
          <div className="flex items-center justify-between p-4 bg-white rounded-lg border">
            <div className="flex items-center gap-3">
              <Zap className="h-5 w-5 text-amber-500" />
              <div>
                <div className="font-medium">Auto-Optimize</div>
                <div className="text-xs text-slate-500">
                  Automatically configure based on dataset size
                </div>
              </div>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={config.autoOptimize}
                onChange={(e) => handleConfigUpdate("autoOptimize", e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-slate-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-500 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>

          {/* Worker Threads */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium flex items-center gap-2">
                <Cpu className="h-4 w-4 text-green-500" />
                Worker Threads
              </label>
              <span className="text-sm font-mono text-slate-600">{config.numWorkers}</span>
            </div>
            <input
              type="range"
              min="8"
              max="48"
              step="8"
              value={config.numWorkers}
              onChange={(e) => handleConfigUpdate("numWorkers", parseInt(e.target.value))}
              disabled={config.autoOptimize}
              className="w-full h-2 bg-slate-200 rounded-lg appearance-none cursor-pointer accent-green-600 disabled:opacity-50 disabled:cursor-not-allowed"
            />
            <div className="flex justify-between text-xs text-slate-500">
              <span>8 (Min)</span>
              <span>24 (Balanced)</span>
              <span>48 (Max)</span>
            </div>
          </div>

          {/* Memory Limit */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium flex items-center gap-2">
                <HardDrive className="h-4 w-4 text-purple-500" />
                Memory Limit
              </label>
              <span className="text-sm font-mono text-slate-600">
                {(config.memoryLimit / 1024).toFixed(1)} GB
              </span>
            </div>
            <input
              type="range"
              min="2048"
              max="16384"
              step="1024"
              value={config.memoryLimit}
              onChange={(e) => handleConfigUpdate("memoryLimit", parseInt(e.target.value))}
              disabled={config.autoOptimize}
              className="w-full h-2 bg-slate-200 rounded-lg appearance-none cursor-pointer accent-purple-600 disabled:opacity-50 disabled:cursor-not-allowed"
            />
            <div className="flex justify-between text-xs text-slate-500">
              <span>2 GB</span>
              <span>8 GB</span>
              <span>16 GB</span>
            </div>
          </div>

          {/* Estimates */}
          <div className="p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-200">
            <div className="flex items-start gap-2 mb-3">
              <Info className="h-4 w-4 text-blue-600 mt-0.5" />
              <div className="text-sm font-semibold text-blue-900">Performance Estimates</div>
            </div>
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div>
                <div className="text-xs text-blue-600 mb-1">Dataset Size</div>
                <div className="font-bold text-blue-900">{datasetSize.toLocaleString()}</div>
              </div>
              <div>
                <div className="text-xs text-blue-600 mb-1">Est. RAM Usage</div>
                <div className="font-bold text-blue-900">{estimatedRAM.toFixed(1)} MB</div>
              </div>
              <div>
                <div className="text-xs text-blue-600 mb-1">Est. Time</div>
                <div className="font-bold text-blue-900">{estimatedTime.toFixed(1)}s</div>
              </div>
            </div>
          </div>

          {/* Warnings */}
          {estimatedRAM > config.memoryLimit && (
            <div className="p-3 bg-amber-50 border border-amber-200 rounded-lg">
              <div className="text-sm font-semibold text-amber-900 mb-1">
                ⚠️ Memory Warning
              </div>
              <div className="text-xs text-amber-700">
                Estimated RAM usage ({estimatedRAM.toFixed(1)} MB) exceeds memory limit. 
                Consider increasing the limit or reducing workers.
              </div>
            </div>
          )}
        </CardContent>
      )}
    </Card>
  );
}
