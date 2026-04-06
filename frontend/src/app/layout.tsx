import type { Metadata } from "next";
import { Rubik, IBM_Plex_Mono } from "next/font/google";
import { AppShell } from "@/components/layout/app-shell";
import { appConfig } from "@/lib/site";
import "./globals.css";

const rubik = Rubik({
  variable: "--font-rubik",
  subsets: ["latin", "cyrillic"],
});

const ibmPlexMono = IBM_Plex_Mono({
  variable: "--font-mono",
  subsets: ["latin", "cyrillic"],
  weight: ["400", "500"],
});

export const metadata: Metadata = {
  title: appConfig.name,
  description: appConfig.description,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ru">
      <body className={`${rubik.variable} ${ibmPlexMono.variable} font-[var(--font-rubik)] antialiased`}>
        <AppShell>{children}</AppShell>
      </body>
    </html>
  );
}
