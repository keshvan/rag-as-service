"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { X } from "lucide-react";
import { navigationItems } from "@/components/layout/nav-items";
import { cn } from "@/lib/cn";

type SidebarProps = {
  isOpen: boolean;
  onClose: () => void;
};

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const pathname = usePathname();

  return (
    <>
      <div
        className={cn(
          "fixed inset-0 z-30 bg-slate-950/45 backdrop-blur-[2px] transition-opacity md:hidden",
          isOpen ? "opacity-100" : "pointer-events-none opacity-0",
        )}
        onClick={onClose}
      />

      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 w-72 border-r border-border/70 bg-panel/95 p-5 shadow-panel backdrop-blur transition-transform md:translate-x-0",
          isOpen ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <div className="mb-8 flex items-center justify-between">
          <Link href="/" className="flex items-center gap-3" onClick={onClose}>
            <Image src="/logo-mark.svg" alt="RAG Консоль" width={36} height={36} priority />
            <div>
              <p className="font-semibold leading-none text-foreground">RAG Консоль</p>
              <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">Пространство</p>
            </div>
          </Link>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md p-1 text-muted transition-colors hover:bg-slate-900/5 hover:text-foreground md:hidden"
            aria-label="Закрыть меню"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <nav className="space-y-2">
          {navigationItems.map((item) => {
            const isActive = pathname === item.href;

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={onClose}
                className={cn(
                  "group block rounded-xl border px-3.5 py-3 transition-all duration-200",
                  isActive
                    ? "border-accent/30 bg-accent/10"
                    : "border-transparent hover:border-border/80 hover:bg-slate-950/5",
                )}
              >
                <div className="mb-1 flex items-center gap-2.5">
                  <item.icon
                    className={cn(
                      "h-4 w-4 transition-transform group-hover:scale-105",
                      isActive ? "text-accent-strong" : "text-muted",
                    )}
                  />
                  <p className={cn("text-sm font-medium", isActive ? "text-foreground" : "text-slate-700")}>
                    {item.label}
                  </p>
                </div>
                <p className="text-xs leading-5 text-muted">{item.description}</p>
              </Link>
            );
          })}
        </nav>
      </aside>
    </>
  );
}
