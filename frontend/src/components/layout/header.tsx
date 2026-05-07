"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Bell, Menu, Search } from "lucide-react";
import { useAuth } from "@/features/auth/providers/auth-provider";
import { resolvePageMeta } from "@/lib/site";

type HeaderProps = {
  onOpenSidebar: () => void;
};

export function Header({ onOpenSidebar }: HeaderProps) {
  const pathname = usePathname();
  const pageMeta = resolvePageMeta(pathname);
  const { user } = useAuth();

  return (
    <header className="sticky top-0 z-20 border-b border-border/70 bg-page/80 backdrop-blur">
      <div className="flex flex-wrap items-center gap-4 px-4 py-4 sm:px-6 lg:px-8">
        <div className="flex min-w-0 flex-1 items-center gap-3">
          <button
            type="button"
            onClick={onOpenSidebar}
            aria-label="Открыть меню"
            className="inline-flex rounded-lg border border-border/70 bg-panel p-2 text-muted transition-colors hover:text-foreground md:hidden"
          >
            <Menu className="h-5 w-5" />
          </button>
          <div className="min-w-0">
            <h1 className="truncate text-xl font-semibold text-foreground">{pageMeta.title}</h1>
            <p className="truncate text-sm text-muted">{pageMeta.subtitle}</p>
          </div>
        </div>

        <div className="flex w-full items-center justify-between gap-3 sm:w-auto sm:justify-end">
          <div className="hidden items-center gap-2 rounded-xl border border-border/70 bg-panel px-3 py-2 text-sm text-muted shadow-sm sm:flex">
            <Search className="h-4 w-4 text-muted" />
            <span>Поиск по документам, задачам и ответам...</span>
          </div>

          <button
            type="button"
            className="inline-flex rounded-xl border border-border/70 bg-panel p-2 text-muted transition-colors hover:text-foreground"
            aria-label="Уведомления"
          >
            <Bell className="h-4 w-4" />
          </button>

          <Link
            href="/profile"
            className="flex items-center gap-3 rounded-xl border border-border/70 bg-panel px-3 py-2 shadow-sm transition-colors hover:border-accent/40 hover:bg-white"
          >
            <div className="h-8 w-8 rounded-full bg-gradient-to-br from-cyan-600 to-amber-400" />
            <div className="leading-tight">
              <p className="text-sm font-medium text-foreground">Профиль</p>
              <p className="text-xs text-muted">{user ? `GUID: ${user.guid}` : "Открыть страницу пользователя"}</p>
            </div>
          </Link>
        </div>
      </div>
    </header>
  );
}
