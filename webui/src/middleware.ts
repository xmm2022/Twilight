import { NextRequest, NextResponse } from "next/server";

/**
 * Middleware: per-request CSP nonce。
 *
 * 旧实现把 `script-src 'self' 'unsafe-inline'` 写死在 next.config.mjs，
 * 反射型 XSS 一旦把 `<script>` 注进 DOM 就直接执行。Next 16 对自己注入的
 * 内联引导脚本不会预先 hash，要么 unsafe-inline，要么 nonce。
 *
 * 这里走 nonce 路线：
 *   1. 每次请求生成一次 16 字节随机 nonce（base64）。
 *   2. 通过请求头 `x-nonce` 传给 RSC，业务代码用 `headers().get('x-nonce')`
 *      给 next/script 设置 nonce 属性；Next 框架本身的内联脚本在响应头里
 *      看到 `nonce-XXX` 后会自动复用同一个值。
 *   3. 响应头里 `script-src 'self' 'nonce-XXX' 'strict-dynamic'`；
 *      `strict-dynamic` 让带 nonce 的脚本动态加载的子脚本继承信任，免去
 *      给每个 _next/static chunk 单独维护 hash。`'self'` 留作 CSP1/2 兜底。
 *   4. 与 next.config.mjs 中静态的安全响应头（X-Frame-Options / HSTS / 等）
 *      共存：那批不依赖请求上下文，留在静态层减少 middleware 开销。
 *
 * 跳过 matcher：静态资源 / 图片 / favicon 等不会嵌脚本，没必要每次重算 nonce。
 */
export function middleware(request: NextRequest) {
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
    `script-src 'self' 'nonce-${nonce}' 'strict-dynamic'${scriptExtras}`,
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
