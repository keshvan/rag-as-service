"use client";

import Link from "next/link";
import type { PropsWithChildren } from "react";

type AuthFormShellProps = PropsWithChildren<{
  eyebrow: string;
  title: string;
  subtitle: string;
  footerText: string;
  footerHref: string;
  footerLabel: string;
}>;

export function AuthFormShell({
  eyebrow,
  title,
  subtitle,
  footerText,
  footerHref,
  footerLabel,
  children,
}: AuthFormShellProps) {
  return (
    <section className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-4 py-10 sm:px-6">
      <div className="grid w-full gap-6 overflow-hidden rounded-[32px] border border-border/70 bg-panel/90 shadow-panel lg:grid-cols-[1.05fr_0.95fr]">
        <div className="relative overflow-hidden bg-[radial-gradient(circle_at_top_left,rgba(14,116,144,0.28),transparent_38%),linear-gradient(135deg,rgba(15,23,42,0.96),rgba(8,102,126,0.92))] p-8 text-white sm:p-10">
          <p className="font-mono text-xs uppercase tracking-[0.28em] text-cyan-100/80">{eyebrow}</p>
          <h1 className="mt-6 max-w-sm text-4xl font-semibold leading-tight">{title}</h1>
          <p className="mt-4 max-w-md text-sm leading-6 text-slate-200">{subtitle}</p>
        </div>

        <div className="p-6 sm:p-10">
          {children}
          <p className="mt-6 text-sm text-muted">
            {footerText}{" "}
            <Link href={footerHref} className="font-medium text-accent-strong transition-colors hover:text-foreground">
              {footerLabel}
            </Link>
          </p>
        </div>
      </div>
    </section>
  );
}
