"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth";
import { useSystemStore } from "@/store/system";
import { Loader2, BookOpen } from "lucide-react";

const landingEnabled = process.env.NEXT_PUBLIC_LANDING_PAGE === "true";

function LandingPage() {
  const { info: systemInfo } = useSystemStore();
  const fetchInfo = useSystemStore((s) => s.fetchInfo);

  useEffect(() => {
    void fetchInfo();
  }, [fetchInfo]);

  // 落地页内容可通过 NEXT_PUBLIC_* 环境变量自定义（构建时注入），
  // 或直接修改下方 JSX。详细见 .env.example 注释。
  const customTitle = process.env.NEXT_PUBLIC_LANDING_TITLE;
  const customDescription = process.env.NEXT_PUBLIC_LANDING_DESCRIPTION;
  const customIcon = process.env.NEXT_PUBLIC_LANDING_ICON;

  const siteName = customTitle || systemInfo?.name || process.env.NEXT_PUBLIC_SITE_NAME || "Twilight";
  const siteDescription = customDescription || (systemInfo?.features
    ? "用户管理面板"
    : "加载中...");
  const iconElement = customIcon
    ? <span className="text-4xl leading-none">{customIcon}</span>
    : <BookOpen className="h-10 w-10 text-primary" />;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-background via-background to-primary/5 p-6">
      <div className="mx-auto max-w-2xl text-center space-y-8">
        <div className="space-y-4">
          <div className="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-primary/10">
            {iconElement}
          </div>
          <h1 className="text-4xl font-bold tracking-tight">{siteName}</h1>
          {siteDescription && (
            <p className="text-lg text-muted-foreground">{siteDescription}</p>
          )}
        </div>

        <div className="flex flex-wrap items-center justify-center gap-4">
          <Link href="/login" prefetch={false}>
            <button className="inline-flex h-11 items-center justify-center rounded-lg bg-primary px-8 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90">
              登录
            </button>
          </Link>
          <Link href="/dashboard" prefetch={false}>
            <button className="inline-flex h-11 items-center justify-center rounded-lg border border-input bg-background px-8 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground">
              进入仪表盘
            </button>
          </Link>
        </div>

        {systemInfo?.version && (
          <div className="pt-8 text-xs text-muted-foreground">
            Version {systemInfo.version}
          </div>
        )}
      </div>
    </div>
  );
}

export default function Home() {
  const router = useRouter();
  const { isAuthenticated, isHydrated, isLoading, initialize } = useAuthStore();

  useEffect(() => {
    if (landingEnabled) return;
    if (!isHydrated) return;
    if (isLoading) {
      void initialize();
      return;
    }
    router.replace(isAuthenticated ? "/dashboard" : "/login");
  }, [isAuthenticated, isHydrated, isLoading, initialize, router]);

  if (landingEnabled) {
    return <LandingPage />;
  }

  return (
    <div className="flex h-screen items-center justify-center">
      <Loader2 className="h-8 w-8 animate-spin text-primary" />
    </div>
  );
}
