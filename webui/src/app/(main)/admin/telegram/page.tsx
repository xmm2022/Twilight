"use client";

import Link from "next/link";
import { Loader2, MessageSquare, RotateCcw, Users } from "lucide-react";
import { AdminConfigSections } from "@/components/admin/config-section-editor";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { api } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { useState } from "react";

export default function AdminTelegramPage() {
  const { toast } = useToast();
  const { t } = useI18n();
  const [testing, setTesting] = useState(false);

  const testBot = async () => {
    setTesting(true);
    try {
      const res = await api.testBotConnectivity();
      toast({
        title: res.success ? t("adminTelegram.testSuccess") : t("adminTelegram.testFailed"),
        description: res.message,
        variant: res.success ? "success" : "destructive",
      });
    } catch (err) {
      toast({ title: t("adminTelegram.testFailed"), description: err instanceof Error ? err.message : undefined, variant: "destructive" });
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">{t("adminTelegram.title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t("adminTelegram.description")}</p>
        </div>
        <Button variant="outline" onClick={() => void testBot()} disabled={testing} className="min-h-10 whitespace-normal leading-tight">
          {testing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RotateCcw className="mr-2 h-4 w-4" />}
          {t("adminTelegram.testBot")}
        </Button>
      </div>

      <div className="grid gap-3 md:grid-cols-2">
        <Link href="/admin/telegram-rebind-requests">
          <Card className="h-full transition-colors hover:border-primary/50 hover:bg-accent/30">
            <CardContent className="flex min-h-[104px] gap-3 p-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <MessageSquare className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <p className="font-medium leading-tight">{t("adminTelegram.rebindTitle")}</p>
                <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{t("adminTelegram.rebindDescription")}</p>
              </div>
            </CardContent>
          </Card>
        </Link>
        <Link href="/admin/scheduler">
          <Card className="h-full transition-colors hover:border-primary/50 hover:bg-accent/30">
            <CardContent className="flex min-h-[104px] gap-3 p-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <Users className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <p className="font-medium leading-tight">{t("adminTelegram.rosterTitle")}</p>
                <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{t("adminTelegram.rosterDescription")}</p>
              </div>
            </CardContent>
          </Card>
        </Link>
      </div>

      <AdminConfigSections
        sectionKeys={["Telegram"]}
        title={t("adminTelegram.configTitle")}
        description={t("adminTelegram.configDescription")}
        notice={t("adminTelegram.configNotice")}
      />
    </div>
  );
}
