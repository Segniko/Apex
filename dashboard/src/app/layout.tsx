import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

import { Providers } from "../components/Providers";

const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || "https://apex-addis.vercel.app";

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: {
    default: "Apex — The Architecture of Recovery",
    template: "%s · Apex",
  },
  description:
    "Industrial-grade crash forensics and observability. Capture, decode, and fix failures with AI-powered root-cause analysis. 100% free and open source.",
  keywords: [
    "crash monitoring",
    "observability",
    "error tracking",
    "AI forensics",
    "open source",
    "Go",
    "Next.js",
    "Sentry alternative",
  ],
  authors: [{ name: "Segniko" }],
  openGraph: {
    type: "website",
    url: SITE_URL,
    siteName: "Apex",
    title: "Apex — The Architecture of Recovery",
    description:
      "Industrial-grade crash forensics with AI root-cause analysis. Capture, decode, and fix failures in real time. 100% free and open source.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Apex — The Architecture of Recovery",
    description:
      "Industrial-grade crash forensics with AI root-cause analysis. 100% free and open source.",
  },
  robots: { index: true, follow: true },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
        suppressHydrationWarning
      >
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
