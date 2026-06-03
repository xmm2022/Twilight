import type { ConfirmOptions } from "@/components/ui/confirm-dialog";
import type { AdminUserListParams } from "@/lib/api-types";

export type BatchDeleteAction = "local_only" | "with_emby";

export interface UsersListState {
  page: number;
  perPage: number;
  search: string;
  roleFilter: string;
  activeFilter: string;
  embyFilter: string;
  sortBy: string;
}

export function buildUsersCacheKey(state: UsersListState): string {
  return [
    state.page,
    state.perPage,
    state.search || "",
    state.roleFilter,
    state.activeFilter,
    state.embyFilter,
    state.sortBy,
  ].join("-");
}

export function usersListParams(state: UsersListState): AdminUserListParams {
  return {
    page: state.page,
    per_page: state.perPage,
    search: state.search || undefined,
    role: state.roleFilter === "all" ? undefined : Number(state.roleFilter),
    active: state.activeFilter === "all" ? undefined : state.activeFilter === "true",
    emby: state.embyFilter === "bound" ? "bound" : state.embyFilter === "unbound" ? "unbound" : undefined,
    sort: state.sortBy || undefined,
  };
}

export function usersBatchFilterParams(state: UsersListState) {
  const params = usersListParams(state);
  return {
    role: params.role,
    active: params.active,
    emby: params.emby,
    search: params.search,
  };
}

export function hasStrongAdminPassword(password: string): boolean {
  return password.length >= 8 && /[a-z]/.test(password) && /[A-Z]/.test(password) && /\d/.test(password);
}

export function toggleSetMember<T>(values: Set<T>, value: T): Set<T> {
  const next = new Set(values);
  if (next.has(value)) next.delete(value);
  else next.add(value);
  return next;
}

export function batchToggleConfirmConfig(enable: boolean, count: number): ConfirmOptions {
  return {
    title: enable ? "启用所选用户？" : "禁用所选用户？",
    description: `将作用于 ${count} 个已选用户。管理员账号由后端保护，会自动跳过。`,
    tone: enable ? "warning" : "danger",
    confirmLabel: enable ? "启用所选" : "禁用所选",
  };
}

export function batchLockEmbyUnbindConfirmConfig(count: number): ConfirmOptions {
  return {
    title: "禁止所选用户自助解绑 Emby？",
    description: `将为 ${count} 个已选目标中已绑定 Emby 的用户写入 Emby 授权锁；未绑定 Emby 的用户会自动跳过。之后用户不能自助解绑 Emby，管理员仍可强制解绑。管理员账号由后端保护。`,
    tone: "warning",
    confirmLabel: "禁止解绑",
  };
}

export function batchDeleteConfirmConfig(count: number, embyCount?: number): ConfirmOptions {
  const embyLabel = typeof embyCount === "number"
    ? `同时删除 Emby 账号（${embyCount} 个）`
    : "同时删除已绑定的 Emby 账号";
  return {
    title: "删除所选用户？",
    description: `将删除 ${count} 个已选用户。管理员账号和当前管理员由后端保护，会自动跳过。`,
    tone: "danger",
    cancelLabel: "取消",
    actions: [
      { label: "仅删除本地账号", value: "local_only" as BatchDeleteAction, variant: "destructive" },
      { label: embyLabel, value: "with_emby" as BatchDeleteAction, variant: "destructive" },
    ],
  };
}
