"use client";

import Link from "next/link";
import type { PropsWithChildren } from "react";

type AuthFormShellProps = PropsWithChildren<{
  title: string;
  footerText: string;
  footerHref: string;
  footerLabel: string;
}>;

export function AuthFormShell({
  title,
  footerText,
  footerHref,
  footerLabel,
  children,
}: AuthFormShellProps) {
  return (
    <section className="flex min-h-screen w-full items-center justify-center px-4 py-8 sm:px-6">
      <div className="w-[500px] max-w-[500px] sm:w-[500px] rounded-2xl border border-border/70 bg-panel p-6 shadow-panel sm:p-8">
        <h1 className="text-center text-2xl font-semibold leading-tight text-foreground">{title}</h1>
        <div className="mt-6">{children}</div>
        <p className="mt-6 text-center text-sm text-muted">
          {footerText}{" "}
          <Link href={footerHref} className="font-medium text-accent-strong transition-colors hover:text-foreground">
            {footerLabel}
          </Link>
        </p>
      </div>
    </section>
  );
}
