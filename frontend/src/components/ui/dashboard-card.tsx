import type { PropsWithChildren } from "react";
import { cn } from "@/lib/cn";

type DashboardCardProps = PropsWithChildren<{
  title: string;
  value: string;
  trend?: string;
  className?: string;
}>;

export function DashboardCard({ title, value, trend, className, children }: DashboardCardProps) {
  return (
    <article className={cn("rounded-2xl border border-border/70 bg-panel p-5 shadow-panel", className)}>
      <p className="text-xs uppercase tracking-[0.16em] text-muted">{title}</p>
      <p className="mt-2 text-3xl font-semibold leading-none text-foreground">{value}</p>
      {trend ? <p className="mt-2 text-sm text-accent-strong">{trend}</p> : null}
      {children ? <div className="mt-4">{children}</div> : null}
    </article>
  );
}
