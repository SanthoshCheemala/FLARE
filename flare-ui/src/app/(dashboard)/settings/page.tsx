"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Settings as SettingsIcon,
  Database,
  Cpu,
  Shield,
  Bell,
  Save,
  RotateCcw,
} from "lucide-react";

export default function SettingsPage() {
  const [settings, setSettings] = useState({
    // PSI Settings
    maxWorkers: 8,
    maxScreenings: 5,
    memoryLimit: 16,

    // Database Settings
    maxConnections: 25,
    connectionTimeout: 30,

    // Security Settings
    sessionTimeout: 900,
    mfaEnabled: false,
    passwordExpiry: 90,

    // Notification Settings
    emailNotifications: true,
    slackNotifications: false,
    webhookUrl: "",
  });

  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    // In a real app, this would call an API
    console.log("Saving settings:", settings);
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  const handleReset = () => {
    setSettings({
      maxWorkers: 8,
      maxScreenings: 5,
      memoryLimit: 16,
      maxConnections: 25,
      connectionTimeout: 30,
      sessionTimeout: 900,
      mfaEnabled: false,
      passwordExpiry: 90,
      emailNotifications: true,
      slackNotifications: false,
      webhookUrl: "",
    });
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
          <p className="text-slate-500">
            Configure system parameters and preferences.
          </p>
          <div className="p-6">
            <h2 className="text-lg font-medium mb-4">System Configuration</h2>
            <div className="space-y-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Sanctions Authority URL</label>
                <input
                  type="text"
                  value="http://localhost:8081"
                  disabled
                  className="flex h-10 w-full rounded-md border border-slate-300 bg-slate-100 px-3 py-2 text-sm ring-offset-white file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-slate-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-950 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                />
                <p className="text-xs text-slate-500">
                  The endpoint of the remote Sanctions Authority server for PSI intersection.
                </p>
              </div>
            </div>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleReset}>
            <RotateCcw className="mr-2 h-4 w-4" />
            Reset
          </Button>
          <Button onClick={handleSave}>
            <Save className="mr-2 h-4 w-4" />
            {saved ? "Saved!" : "Save Changes"}
          </Button>
        </div>
      </div>

      {/* PSI Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Cpu className="h-5 w-5" />
            PSI Engine Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                Max Worker Threads
              </label>
              <input
                type="number"
                value={settings.maxWorkers}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    maxWorkers: parseInt(e.target.value),
                  })
                }
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-slate-500">
                Number of CPU cores to use for parallel processing (1-
                {typeof navigator !== "undefined" ? navigator.hardwareConcurrency || 8 : 8})
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                Max Concurrent Screenings
              </label>
              <input
                type="number"
                value={settings.maxScreenings}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    maxScreenings: parseInt(e.target.value),
                  })
                }
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-slate-500">
                Maximum number of screenings that can run simultaneously
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                Memory Limit (GB)
              </label>
              <input
                type="number"
                value={settings.memoryLimit}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    memoryLimit: parseInt(e.target.value),
                  })
                }
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-slate-500">
                Maximum RAM allocation for PSI operations
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Database Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            Database Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                Max Connections
              </label>
              <input
                type="number"
                value={settings.maxConnections}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    maxConnections: parseInt(e.target.value),
                  })
                }
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-slate-500">
                Maximum database connection pool size
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                Connection Timeout (seconds)
              </label>
              <input
                type="number"
                value={settings.connectionTimeout}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    connectionTimeout: parseInt(e.target.value),
                  })
                }
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-slate-500">
                Database connection timeout duration
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Security Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Security Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                Session Timeout (seconds)
              </label>
              <input
                type="number"
                value={settings.sessionTimeout}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    sessionTimeout: parseInt(e.target.value),
                  })
                }
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-slate-500">
                User session expiration time
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">
                Password Expiry (days)
              </label>
              <input
                type="number"
                value={settings.passwordExpiry}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    passwordExpiry: parseInt(e.target.value),
                  })
                }
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-slate-500">
                Days until password must be changed
              </p>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="mfa"
                checked={settings.mfaEnabled}
                onChange={(e) =>
                  setSettings({ ...settings, mfaEnabled: e.target.checked })
                }
                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
              />
              <label
                htmlFor="mfa"
                className="text-sm font-medium text-slate-700"
              >
                Enable Multi-Factor Authentication
              </label>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Notification Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5" />
            Notification Settings
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="email"
                checked={settings.emailNotifications}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    emailNotifications: e.target.checked,
                  })
                }
                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
              />
              <label
                htmlFor="email"
                className="text-sm font-medium text-slate-700"
              >
                Email Notifications
              </label>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="slack"
                checked={settings.slackNotifications}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    slackNotifications: e.target.checked,
                  })
                }
                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
              />
              <label
                htmlFor="slack"
                className="text-sm font-medium text-slate-700"
              >
                Slack Notifications
              </label>
            </div>

            {settings.slackNotifications && (
              <div className="space-y-2 ml-6">
                <label className="text-sm font-medium text-slate-700">
                  Webhook URL
                </label>
                <input
                  type="url"
                  value={settings.webhookUrl}
                  onChange={(e) =>
                    setSettings({ ...settings, webhookUrl: e.target.value })
                  }
                  placeholder="https://hooks.slack.com/services/..."
                  className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* System Information */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <SettingsIcon className="h-5 w-5" />
            System Information
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <p className="text-sm font-medium text-slate-700">Version</p>
              <p className="text-sm text-slate-600">1.0.0</p>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-700">Environment</p>
              <p className="text-sm text-slate-600">Development</p>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-700">Database</p>
              <p className="text-sm text-slate-600">SQLite</p>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-700">PSI Library</p>
              <p className="text-sm text-slate-600">LE-PSI v0.1.0</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
