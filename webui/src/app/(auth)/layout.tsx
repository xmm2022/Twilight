"use client";

import { useEffect } from "react";
import { useSystemStore } from "@/store/system";
import { sanitizeImageUrl } from "@/lib/safe-url";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { info: systemInfo, fetchInfo } = useSystemStore();
  const bgUrl = systemInfo?.auth_background_url;
  const safeBg = bgUrl ? sanitizeImageUrl(bgUrl) : undefined;

  useEffect(() => {
    void fetchInfo();
  }, [fetchInfo]);

  const backgroundStyle = safeBg
    ? {
        backgroundImage: `url(${safeBg})`,
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundRepeat: "no-repeat",
      }
    : undefined;

  return (
    <div
      className="relative min-h-screen overflow-hidden bg-background"
      style={backgroundStyle}
    >
      {!safeBg && (
        <>
          <div className="shell-glow shell-glow-left" />
          <div className="shell-glow shell-glow-right" />
          <div className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_20%_10%,hsl(var(--primary)/0.12),transparent_35%),radial-gradient(circle_at_80%_90%,hsl(var(--primary)/0.08),transparent_30%)]" />
        </>
      )}
      <div className="relative z-10 min-h-screen">
        {children}
      </div>
    </div>
  );
}
