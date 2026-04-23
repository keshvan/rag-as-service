"use client";

import { useEffect, useState, type PropsWithChildren } from "react";
import { usePathname } from "next/navigation";
import { Header } from "@/components/layout/header";
import { Sidebar } from "@/components/layout/sidebar";

export function AppShell({ children }: PropsWithChildren) {
  const pathname = usePathname();
  const [isSidebarOpen, setSidebarOpen] = useState(false);
  const isAuthPage = pathname === "/login" || pathname === "/register" || pathname === "/verification";

  useEffect(() => {
    const onResize = () => {
      if (window.innerWidth >= 768) {
        setSidebarOpen(false);
      }
    };

    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  if (isAuthPage) {
    return <div className="min-h-screen bg-page text-foreground">{children}</div>;
  }

  return (
    <div className="min-h-screen bg-page text-foreground">
      <Sidebar isOpen={isSidebarOpen} onClose={() => setSidebarOpen(false)} />
      <div className="min-h-screen md:pl-72">
        <Header onOpenSidebar={() => setSidebarOpen(true)} />
        <main className="px-4 pb-8 pt-6 sm:px-6 lg:px-8">{children}</main>
      </div>
    </div>
  );
}
