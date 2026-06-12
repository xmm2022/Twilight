"use client";

import { useCallback, useState } from "react";
import {
  Trash2,
  Loader2,
  ChevronLeft,
  ChevronRight,
  Search,
  ScrollText,
  User,
  Shield,
  Bot,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";
import { useAsyncResource } from "@/hooks/use-async-resource";
import { PageError } from "@/components/layout/page-state";
import { api, type AuditLog } from "@/lib/api";
import { formatDate } from "@/lib/utils";
import { useI18n, type MessageKey } from "@/lib/i18n";

const CATEGORY_ICONS: Record<string, React.ReactNode> = {
  admin: <Shield className="h-4 w-4" />,
  user: <User className="h-4 w-4" />,
  system: <Bot className="h-4 w-4" />,
};

const ACTION_LABELS: Record<string, MessageKey> = {
  create_regcode: "adminAuditLog.actionCreateRegcode",
  update_regcode: "adminAuditLog.actionUpdateRegcode",
  delete_regcode: "adminAuditLog.actionDeleteRegcode",
  batch_delete_regcode: "adminAuditLog.actionBatchDeleteRegcode",
  clear_regcode_usage: "adminAuditLog.actionClearRegcodeUsage",
  create_invite_code: "adminAuditLog.actionCreateInviteCode",
  create_renew_code: "adminAuditLog.actionCreateRenewCode",
  use_code: "adminAuditLog.actionUseCode",
  update_user: "adminAuditLog.actionUpdateUser",
  set_role: "adminAuditLog.actionSetRole",
  enable_user: "adminAuditLog.actionEnableUser",
  disable_user: "adminAuditLog.actionDisableUser",
  delete_user: "adminAuditLog.actionDeleteUser",
  batch_enable_users: "adminAuditLog.actionBatchEnableUsers",
  batch_disable_users: "adminAuditLog.actionBatchDisableUsers",
  batch_renew_users: "adminAuditLog.actionBatchRenewUsers",
  batch_delete_users: "adminAuditLog.actionBatchDeleteUsers",
};

export default function AdminAuditLogsPage() {
  const { toast } = useToast();
  const { t } = useI18n();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [actionFilter, setActionFilter] = useState("all");
  const [search, setSearch] = useState("");
  const [searchInput, setSearchInput] = useState("");
  const [clearOpen, setClearOpen] = useState(false);
  const [isClearing, setIsClearing] = useState(false);
  const perPage = 50;

  const loadLogs = useCallback(
    async (signal?: AbortSignal) => {
      const res = await api.getAuditLogs(page, {
        category: categoryFilter !== "all" ? categoryFilter : undefined,
        action: actionFilter !== "all" ? actionFilter : undefined,
        search: search || undefined,
        per_page: perPage,
        signal,
      });
      if (res.success && res.data) {
        setLogs(res.data.logs || []);
        setTotal(res.data.total || 0);
      }
      return true;
    },
    [page, categoryFilter, actionFilter, search]
  );

  const { isLoading, error, execute: reload } = useAsyncResource(
    loadLogs,
    { immediate: true }
  );

  const handleDelete = async (id: number) => {
    try {
      const res = await api.deleteAuditLog(id);
      if (res.success) {
        toast({ title: t("adminAuditLog.deleted"), variant: "success" });
        reload();
      } else {
        toast({ title: t("common.deleteFailed"), description: res.message, variant: "destructive" });
      }
    } catch (err: unknown) {
      toast({ title: t("common.deleteFailed"), description: (err as Error).message, variant: "destructive" });
    }
  };

  const handleClearAll = async () => {
    setIsClearing(true);
    try {
      const res = await api.clearAuditLogs();
      if (res.success) {
        toast({ title: t("adminAuditLog.clearedAll"), variant: "success" });
        setClearOpen(false);
        reload();
      } else {
        toast({ title: t("adminAuditLog.clearFailed"), description: res.message, variant: "destructive" });
      }
    } catch (err: unknown) {
      toast({ title: t("adminAuditLog.clearFailed"), description: (err as Error).message, variant: "destructive" });
    } finally {
      setIsClearing(false);
    }
  };

  const handleSearch = () => {
    setSearch(searchInput);
    setPage(1);
  };

  const totalPages = Math.ceil(total / perPage);

  if (error) return <PageError message={error} />;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ScrollText className="h-6 w-6 text-primary" />
          <h1 className="text-2xl font-bold">{t("adminAuditLog.title")}</h1>
          {total > 0 && (
            <Badge variant="secondary" className="ml-2">
              {total}
            </Badge>
          )}
        </div>
        {total > 0 && (
          <Button
            variant="destructive"
            size="sm"
            onClick={() => setClearOpen(true)}
          >
            {t("adminAuditLog.clearAll")}
          </Button>
        )}
      </div>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">{t("adminAuditLog.filter")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col sm:flex-row gap-3 flex-wrap">
            <Select value={categoryFilter} onValueChange={(v) => { setCategoryFilter(v); setPage(1); }}>
              <SelectTrigger className="w-full sm:w-[160px]">
                <SelectValue placeholder={t("adminAuditLog.filterCategory")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t("adminAuditLog.allCategories")}</SelectItem>
                <SelectItem value="admin">{t("adminAuditLog.categoryAdmin")}</SelectItem>
                <SelectItem value="user">{t("adminAuditLog.categoryUser")}</SelectItem>
                <SelectItem value="system">{t("adminAuditLog.categorySystem")}</SelectItem>
              </SelectContent>
            </Select>
            <Select value={actionFilter} onValueChange={(v) => { setActionFilter(v); setPage(1); }}>
              <SelectTrigger className="w-full sm:w-[200px]">
                <SelectValue placeholder={t("adminAuditLog.filterAction")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t("adminAuditLog.allActions")}</SelectItem>
                <SelectItem value="create_regcode">{t("adminAuditLog.actionCreateRegcode")}</SelectItem>
                <SelectItem value="update_regcode">{t("adminAuditLog.actionUpdateRegcode")}</SelectItem>
                <SelectItem value="delete_regcode">{t("adminAuditLog.actionDeleteRegcode")}</SelectItem>
                <SelectItem value="batch_delete_regcode">{t("adminAuditLog.actionBatchDeleteRegcode")}</SelectItem>
                <SelectItem value="clear_regcode_usage">{t("adminAuditLog.actionClearRegcodeUsage")}</SelectItem>
                <SelectItem value="create_invite_code">{t("adminAuditLog.actionCreateInviteCode")}</SelectItem>
                <SelectItem value="create_renew_code">{t("adminAuditLog.actionCreateRenewCode")}</SelectItem>
                <SelectItem value="use_code">{t("adminAuditLog.actionUseCode")}</SelectItem>
                <SelectItem value="update_user">{t("adminAuditLog.actionUpdateUser")}</SelectItem>
                <SelectItem value="set_role">{t("adminAuditLog.actionSetRole")}</SelectItem>
                <SelectItem value="enable_user">{t("adminAuditLog.actionEnableUser")}</SelectItem>
                <SelectItem value="disable_user">{t("adminAuditLog.actionDisableUser")}</SelectItem>
                <SelectItem value="delete_user">{t("adminAuditLog.actionDeleteUser")}</SelectItem>
                <SelectItem value="batch_enable_users">{t("adminAuditLog.actionBatchEnableUsers")}</SelectItem>
                <SelectItem value="batch_disable_users">{t("adminAuditLog.actionBatchDisableUsers")}</SelectItem>
                <SelectItem value="batch_renew_users">{t("adminAuditLog.actionBatchRenewUsers")}</SelectItem>
                <SelectItem value="batch_delete_users">{t("adminAuditLog.actionBatchDeleteUsers")}</SelectItem>
              </SelectContent>
            </Select>
            <div className="flex gap-2 flex-1 min-w-[200px]">
              <Input
                placeholder={t("adminAuditLog.searchPlaceholder")}
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              />
              <Button variant="outline" size="icon" onClick={handleSearch}>
                <Search className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : logs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
              <ScrollText className="h-10 w-10 mb-3 opacity-40" />
              <p>{t("adminAuditLog.empty")}</p>
            </div>
          ) : (
            <div className="divide-y">
              {logs.map((log) => (
                <div
                  key={log.id}
                  className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4 p-4"
                >
                  <div className="flex-1 min-w-0 space-y-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      {CATEGORY_ICONS[log.category] && (
                        <span className="text-muted-foreground">
                          {CATEGORY_ICONS[log.category]}
                        </span>
                      )}
                      <span className="font-medium">{log.username}</span>
                      <Badge variant="outline" className="text-xs">
                        UID: {log.uid}
                      </Badge>
                      <Badge
                        variant={log.category === "admin" ? "default" : log.category === "system" ? "secondary" : "outline"}
                        className="text-xs"
                      >
                        {log.category === "admin"
                          ? t("adminAuditLog.categoryAdmin")
                          : log.category === "system"
                            ? t("adminAuditLog.categorySystem")
                            : t("adminAuditLog.categoryUser")}
                      </Badge>
                    </div>
                    <div className="text-sm">
                      <span className="font-medium">
                        {ACTION_LABELS[log.action] ? t(ACTION_LABELS[log.action]) : log.action}
                      </span>
                      {log.target_uid != null && log.target_uid > 0 && (
                        <span className="ml-2 text-muted-foreground">
                          → UID: {log.target_uid}
                        </span>
                      )}
                    </div>
                    {log.detail && Object.keys(log.detail).length > 0 && (
                      <div className="text-xs text-muted-foreground">
                        {JSON.stringify(log.detail)}
                      </div>
                    )}
                    {log.ip && (
                      <div className="text-xs text-muted-foreground">
                        IP: {log.ip}
                      </div>
                    )}
                  </div>
                  <div className="flex items-center gap-3 sm:flex-shrink-0">
                    <span className="text-xs text-muted-foreground whitespace-nowrap">
                      {formatDate(log.created_at)}
                    </span>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => handleDelete(log.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage(page - 1)}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-muted-foreground">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage(page + 1)}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      <Dialog open={clearOpen} onOpenChange={setClearOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("adminAuditLog.clearConfirmTitle")}</DialogTitle>
            <DialogDescription>
              {t("adminAuditLog.clearConfirmDescription", { count: total })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setClearOpen(false)}>
              {t("common.cancel")}
            </Button>
            <Button
              variant="destructive"
              onClick={handleClearAll}
              disabled={isClearing}
            >
              {isClearing && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              {t("adminAuditLog.clearConfirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
