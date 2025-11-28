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
      icon: History, 
      label: "Audit Log", 
      href: "/audit",
      show: true 
    },
    { 
      icon: Settings, 
      label: "Settings", 
      href: "/settings",
      show: true 
    },
  ];

  return (
    <aside className="hidden w-64 flex-col border-r bg-slate-50/40 md:flex h-screen sticky top-0">
      <div className="p-6 border-b">
        <h1 className="text-xl font-bold tracking-tight text-slate-900">
          {isClientMode() ? "FLARE Bank Node" : "FLARE Authority"}
        </h1>
        <p className="text-xs text-slate-500 mt-1">
          {isClientMode() ? "Client Portal" : "Server Administration"}
        </p>
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
