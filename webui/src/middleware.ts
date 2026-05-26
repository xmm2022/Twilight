import { NextRequest, NextResponse } from "next/server";

/**
 * Middleware：per-request CSP nonce + 服务端 session cookie 守卫。
 *
 * 两件事合在一起做的原因是 Next 的 middleware 是请求最早的 server-side hook，
 * 想避开"未登录用户先看到管理面板再被 client-side router.push('/login') 踢
 * 走"的肉眼闪烁，必须在 RSC payload 写出之前就完成 redirect——而 RSC payload
 * 是由这个 middleware 之后的 React 服务器渲染产出的。
 *
 * ## CSP nonce
 *
 * 旧实现 `script-src 'self' 'unsafe-inline'` 写死在 next.config.mjs，反射型
 * XSS 一旦把 `<script>` 注进 DOM 就直接执行。Next 16 对自己注入的内联引导
 * 脚本不会预先 hash，要么 unsafe-inline，要么 nonce。这里走 nonce 路线：
 *
 *   1. 每次请求生成 16 字节随机 nonce（base64）。
 *   2. 通过请求头 `x-nonce` 传给 RSC，业务代码用 `headers().get('x-nonce')`
 *      给 next/script 设置 nonce 属性；Next 框架本身的内联脚本在响应头里
 *      看到 `nonce-XXX` 后会自动复用同一个值。
 *   3. 响应头里 `script-src 'self' 'nonce-XXX'`。
 *      历史上这里曾叠加 `'strict-dynamic'`，理论收益是"带 nonce 的脚本动
 *      态加载的子脚本自动继承信任，免去给每个 _next/static chunk 单独维
 *      护 hash"。但 `'strict-dynamic'` 一旦出现就会让浏览器忽略 `'self'`
 *      —— Next 16 的 auto-nonce 注入对 RSC payload 里 parser 插入的
 *      `<script src="/_next/static/chunks/...">` 偶尔漏标 nonce（已知
 *      vercel/next.js issue），生产环境就会出现一条 chunk 因没有 nonce
 *      被整段拒绝，整页白屏。
 *
 *      去掉 `'strict-dynamic'` 后：内联 bootstrap 仍走 nonce 通过；同源
 *      chunk 走 `'self'` 通过；XSS 表面跟之前一样——本应用从来不通过
 *      `<script src=>` 加载第三方脚本，`'self'` 不比 `'strict-dynamic'`
 *      宽松。
 *   4. 与 next.config.mjs 中静态的安全响应头（X-Frame-Options / HSTS / 等）
 *      共存：那批不依赖请求上下文，留在静态层减少 middleware 开销。
 *
 * ## Session cookie 守卫
 *
 * 后端在登录成功时写 `twilight_session`（HttpOnly + SameSite=Lax）。客户端
 * JS 拿不到，但 middleware 在 server 端可以读：
 *
 *   - protectedPrefixes 里的路径若没有 cookie ⇒ 302 -> /login?next=<path>
 *   - authPrefixes 里的路径若已有 cookie ⇒ 302 -> /dashboard
 *
 * 注意：cookie 仅证明"曾经登录过"，session 是否真的有效仍由后端在每个 API
 * 请求里校验；这里的目的只是消除 SSR 阶段的"先渲染管理面板，再被 client
 * effect 踢走"闪烁，不替代后端鉴权。前端的 router 仍然保留 useEffect 兜底，
 * 应对"cookie 还在但被 server 标记 invalid / 5xx 后清掉"的退化场景。
 */

const protectedPrefixes = [
  "/dashboard",
  "/admin",
  "/announcements",
  "/invite",
  "/media",
  "/score",
  "/settings",
];

const authPrefixes = ["/login", "/register", "/forgot-password"];

const SESSION_COOKIE = "twilight_session";

function pathMatches(pathname: string, prefixes: string[]): boolean {
  for (const prefix of prefixes) {
    if (pathname === prefix || pathname.startsWith(prefix + "/")) {
      return true;
    }
  }
  return false;
}

export function middleware(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  const hasSession = Boolean(request.cookies.get(SESSION_COOKIE)?.value);

  // 1) 未登录访问受保护页面 → 直接 302 到 /login？next=...
  //    用 redirect 而不是 rewrite：浏览器地址栏要变成 /login，避免用户在
  //    一个看起来仍在 /admin 的 URL 上看到登录页（既会让书签错乱，也容易
  //    被钓鱼站借用）。`next` 仅在白名单内才回填，避免 open redirect。
  if (!hasSession && pathMatches(pathname, protectedPrefixes)) {
    const loginURL = new URL("/login", request.url);
    if (pathMatches(pathname, protectedPrefixes)) {
      loginURL.searchParams.set("next", pathname + (search || ""));
    }
    return NextResponse.redirect(loginURL);
  }

  // 2) 已登录访问登录/注册/找回密码 → 直接送回 /dashboard，避免回退按钮
  //    把已登录用户卡在登录页上反复 submit。
  if (hasSession && pathMatches(pathname, authPrefixes)) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  // 3) 走到这里说明鉴权 OK，继续做 CSP nonce 注入。
  // 16 字节足够抵御穷举（128 bit），编码后 24 字符。
  const random = new Uint8Array(16);
  crypto.getRandomValues(random);
  let raw = "";
  for (const byte of random) raw += String.fromCharCode(byte);
  const nonce = btoa(raw);

  const isDev = process.env.NODE_ENV !== "production";
  // dev 下 Next 用 eval 做 HMR / RSC payload 解析；生产构建后丢掉 unsafe-eval。
  const scriptExtras = isDev ? " 'unsafe-eval'" : "";

  const extraConnect = process.env.NEXT_PUBLIC_CSP_CONNECT?.trim();
  const connectSrc = extraConnect ? `'self' ${extraConnect}` : "'self'";

  const csp = [
    "default-src 'self'",
    `script-src 'self' 'nonce-${nonce}'${scriptExtras}`,
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data: blob: https:",
    "font-src 'self' data:",
    `connect-src ${connectSrc}`,
    "media-src 'self' https:",
    "frame-ancestors 'none'",
    "object-src 'none'",
    "base-uri 'self'",
    "form-action 'self'",
    "worker-src 'self' blob:",
    "manifest-src 'self'",
    "upgrade-insecure-requests",
  ].join("; ");

  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-nonce", nonce);
  // Next 框架读响应头里的 CSP 自动给自己注入的 inline script 加 nonce；
  // 业务自己写 next/script 时也走 headers().get('x-nonce') 显式标。
  requestHeaders.set("content-security-policy", csp);

  const response = NextResponse.next({ request: { headers: requestHeaders } });
  response.headers.set("Content-Security-Policy", csp);
  return response;
}

export const config = {
  matcher: [
    // 跳过 _next 静态资源、图片优化、favicon、其它静态资产；
    // 这些请求永远不会触发脚本执行，跳过 middleware 节省每请求 CPU。
    "/((?!_next/static|_next/image|favicon\\.ico|favicon\\.svg|images/|api/).*)",
  ],
};
