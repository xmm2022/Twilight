import { redirect } from "next/navigation";

/**
 * Root path 不再读取 cookie 猜测登录态，统一进入 /dashboard，由 main layout
 * 调 `/users/me` 判定是否需要跳回 /login。
 */
export default function Home() {
  redirect("/dashboard");
}
