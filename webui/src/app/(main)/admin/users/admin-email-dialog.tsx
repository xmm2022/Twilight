"use client";

import { useEffect, useState } from "react";
import { Loader2, Mail } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { useI18n } from "@/lib/i18n";
import { api, type UserInfo } from "@/lib/api";
import { friendlyError } from "@/lib/validators";

// AdminEmailDialog 是自包含的「邮箱验证管理」对话框：管理员可强制把用户绑定到
// 指定邮箱（默认直接标记已验证），或在不改邮箱的前提下置/撤销验证状态。
// 自带状态与请求，page.tsx 只需提供 user + open + onDone（刷新列表）。
export function AdminEmailDialog({
  open,
  onOpenChange,
  user,
  onDone,
}: {
  open: boolean;
  onOpenChange: (next: boolean) => void;
  user: UserInfo | null;
  onDone: () => void;
}) {
  const { t } = useI18n();
  const { toast } = useToast();
  const [email, setEmail] = useState("");
  const [markVerified, setMarkVerified] = useState(true);
  const [force, setForce] = useState(false);
  const [loading, setLoading] = useState(false);

  // 每次打开同步当前用户邮箱与默认选项。
  useEffect(() => {
    if (open) {
      setEmail(user?.email || "");
      setMarkVerified(true);
      setForce(false);
    }
  }, [open, user?.uid, user?.email]);

  if (!user) return null;

  const verified = Boolean(user.email_verified);
  const hasEmail = Boolean(user.email);

  const bind = async () => {
    if (!email.trim()) {
      toast({ title: t("email.enterEmailFirst"), variant: "destructive" });
      return;
    }
    setLoading(true);
    try {
      const res = await api.adminBindUserEmail(user.uid, { email: email.trim(), mark_verified: markVerified, force });
      if (res.success) {
        toast({ title: t("email.admin.bindSuccess"), variant: "success" });
        onDone();
        onOpenChange(false);
      } else {
        toast({ title: friendlyError(res.error_code, res.message), variant: "destructive" });
      }
    } catch (error: any) {
      toast({ title: error?.message || t("common.networkError"), variant: "destructive" });
    } finally {
      setLoading(false);
    }
  };

  const setVerified = async (next: boolean) => {
    setLoading(true);
    try {
      const res = await api.adminSetUserEmailVerified(user.uid, { verified: next, force: true });
      if (res.success) {
        toast({ title: t("email.admin.verifiedUpdated"), variant: "success" });
        onDone();
        onOpenChange(false);
      } else {
        toast({ title: friendlyError(res.error_code, res.message), variant: "destructive" });
      }
    } catch (error: any) {
      toast({ title: error?.message || t("common.networkError"), variant: "destructive" });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            {t("email.admin.bindTitle")}
          </DialogTitle>
          <DialogDescription>
            {user.username} · {t("email.admin.bindDescription")}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-sm">
            <span className="text-muted-foreground">{t("email.currentEmail")}:</span>
            <span className="font-medium">{user.email || t("email.notBound")}</span>
            {hasEmail && (
              <Badge variant="outline" className={verified ? "border-emerald-500/40 text-emerald-600" : "border-amber-500/40 text-amber-600"}>
                {verified ? t("email.verified") : t("email.unverified")}
              </Badge>
            )}
          </div>
          <div className="space-y-2">
            <Label>{t("email.emailLabel")}</Label>
            <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder={t("email.emailPlaceholder")} />
          </div>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input type="checkbox" className="h-4 w-4 rounded" checked={markVerified} onChange={(e) => setMarkVerified(e.target.checked)} />
            {t("email.admin.markVerified")}
          </label>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input type="checkbox" className="h-4 w-4 rounded" checked={force} onChange={(e) => setForce(e.target.checked)} />
            {t("email.admin.force")}
          </label>
          {hasEmail && (
            <div className="flex gap-2 border-t pt-3">
              {verified ? (
                <Button type="button" variant="outline" size="sm" disabled={loading} onClick={() => setVerified(false)}>
                  {t("email.admin.unsetVerified")}
                </Button>
              ) : (
                <Button type="button" variant="outline" size="sm" disabled={loading} onClick={() => setVerified(true)}>
                  {t("email.admin.setVerified")}
                </Button>
              )}
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t("common.cancel")}</Button>
          <Button onClick={bind} disabled={loading}>
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {t("email.admin.bindAction")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
