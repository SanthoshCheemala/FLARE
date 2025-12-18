"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  Users,
  ShieldAlert,
  ScanSearch,
  History,
  Settings,
  FileText,
  Activity,
  Flame,
} from "lucide-react";
import { isClientMode, isServerMode } from "@/lib/config";

export function Sidebar() {
  const pathname = usePathname();

  const sidebarItems = [
    { 
      icon: LayoutDashboard, 
      label: "Dashboard", 
      href: "/dashboard",
      show: true 
    },
    { 
      icon: Users, 
      label: "Customers", 
      href: "/customers",
      show: isClientMode() 
    },
    { 
      icon: ShieldAlert, 
      label: isServerMode() ? "Manage Sanctions" : "Sanctions Lists", 
      href: "/sanctions",
      show: true 
    },
    { 
      icon: ScanSearch, 
      label: "Run Screening", 
      href: "/screening",
      show: isClientMode() 
    },
    { 
      icon: FileText, 
      label: "Results", 
      href: "/results",
      show: isClientMode() 
    },
    { 
      icon: Activity, 
      label: "Performance", 
      href: "/performance",
      show: isClientMode() 
    },
    { 
      icon: History, 
      label: "Audit Log", 
      href: "/audit",
      show: true 
    },
    { 
      icon: Settings, 
      label: "Settings", 
      href: "/settings",
      show: false // Hidden from navigation
    },
  ];

  return (
    <aside className="hidden w-64 flex-col border-r bg-gradient-to-b from-slate-50 to-slate-100/50 md:flex h-screen sticky top-0">
      {/* Accent Bar */}
      <div className="h-1 bg-gradient-to-r from-orange-500 via-red-500 to-pink-500" />
      
      {/* Logo & Title */}
      <div className="p-6 border-b bg-white/50">
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-orange-500 to-red-600 shadow-lg shadow-orange-500/25">
            <Flame className="h-5 w-5 text-white" />
          </div>
          <div>
            <h1 className="text-lg font-bold tracking-tight bg-gradient-to-r from-slate-800 to-slate-600 bg-clip-text text-transparent">
              {isClientMode() ? "FLARE" : "FLARE"}
            </h1>
            <p className="text-[10px] font-medium text-slate-500 uppercase tracking-wider">
              {isClientMode() ? "Privacy Screening" : "Sanctions Authority"}
            </p>
          </div>
        </div>
      </div>
      <nav className="flex-1 overflow-y-auto p-4 space-y-1">
        {sidebarItems.filter(item => item.show).map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
              pathname === item.href || pathname?.startsWith(item.href + "/")
                ? "bg-slate-900 text-slate-50"
                : "text-slate-700 hover:bg-slate-100 hover:text-slate-900"
            )}
          >
            <item.icon className="h-4 w-4" />
            {item.label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}
