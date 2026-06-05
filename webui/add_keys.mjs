import fs from "fs";

const additions = {
  basic: {
    common: {
      permanent: "永久",
      notActivated: "未开通",
      invalidDate: "无效日期",
      relativeTime: {
        expiredAgo: "已过期 {days} 天",
        today: "今天到期",
        tomorrow: "明天到期",
        inDays: "{days} 天后到期",
        inWeeks: "{weeks} 周后到期",
        inMonths: "{months} 个月后到期",
        inYears: "{years} 年后到期",
      },
      httpError: {
        unauthorized: "登录态已失效，请重新登录",
        forbidden: "权限不足",
        notFound: "请求的资源不存在",
        conflict: "操作冲突，请刷新后重试",
        payloadTooLarge: "上传内容过大",
        tooManyRequests: "请求过于频繁，请稍后再试",
        serverError: "服务器开小差了，请稍后再试",
        requestFailed: "请求失败",
      },
    },
    errorBoundary: { errorId: "错误 ID" },
    validators: {
      usernameRequired: "请填写用户名",
      usernameLength: "用户名长度需为 {min}-{max} 位",
      usernameFormat: "用户名格式不正确：仅允许字母 / 数字 / 下划线，且不能以数字开头",
      embyUsernameRequired: "请填写 Emby 用户名",
      embyUsernameTooLong: "Emby 用户名长度不能超过 {max} 位",
      embyUsernameControlChars: "Emby 用户名包含不可见字符",
      emailTooLong: "邮箱长度不能超过 254 位",
      emailFormat: "邮箱格式不正确",
      codeRequired: "请填写注册码 / 邀请码",
      codeLength: "注册码长度需为 {min}-{max} 位",
      codeFormat: "注册码只允许字母 / 数字 / 下划线 / 短横线",
      intMustBeInteger: "{label} 必须是整数",
      intMin: "{label} 不能小于 {min}",
      intMax: "{label} 不能大于 {max}",
    },
    password: {
      defaultLabel: "新密码",
      required: "请提供{label}",
      tooLong: "{label}过长，最多 128 位",
      tooWeak: "{label}强度不足：至少 8 位，且包含大小写字母和数字",
      needLower: "{label}强度不足：至少包含一个小写字母",
      needUpper: "{label}强度不足：至少包含一个大写字母",
      needDigit: "{label}强度不足：至少包含一个数字",
      ok: "强度合格",
      levelWeak: "弱",
      levelFair: "一般",
      levelGood: "良好",
      levelStrong: "强",
    },
  },

  "zh-Hant": {
    common: {
      permanent: "永久",
      notActivated: "未開通",
      invalidDate: "無效日期",
      relativeTime: {
        expiredAgo: "已過期 {days} 天",
        today: "今天到期",
        tomorrow: "明天到期",
        inDays: "{days} 天後到期",
        inWeeks: "{weeks} 週後到期",
        inMonths: "{months} 個月後到期",
        inYears: "{years} 年後到期",
      },
      httpError: {
        unauthorized: "登入狀態已失效，請重新登入",
        forbidden: "權限不足",
        notFound: "請求的資源不存在",
        conflict: "操作衝突，請重新整理後重試",
        payloadTooLarge: "上傳內容過大",
        tooManyRequests: "請求過於頻繁，請稍後再試",
        serverError: "伺服器出了點問題，請稍後再試",
        requestFailed: "請求失敗",
      },
    },
    errorBoundary: { errorId: "錯誤 ID" },
    validators: {
      usernameRequired: "請填寫使用者名稱",
      usernameLength: "使用者名稱長度需為 {min}-{max} 位",
      usernameFormat: "使用者名稱格式不正確：僅允許字母 / 數字 / 底線，且不能以數字開頭",
      embyUsernameRequired: "請填寫 Emby 使用者名稱",
      embyUsernameTooLong: "Emby 使用者名稱長度不能超過 {max} 位",
      embyUsernameControlChars: "Emby 使用者名稱包含不可見字元",
      emailTooLong: "信箱長度不能超過 254 位",
      emailFormat: "信箱格式不正確",
      codeRequired: "請填寫註冊碼 / 邀請碼",
      codeLength: "註冊碼長度需為 {min}-{max} 位",
      codeFormat: "註冊碼只允許字母 / 數字 / 底線 / 連字號",
      intMustBeInteger: "{label} 必須是整數",
      intMin: "{label} 不能小於 {min}",
      intMax: "{label} 不能大於 {max}",
    },
    password: {
      defaultLabel: "新密碼",
      required: "請提供{label}",
      tooLong: "{label}過長，最多 128 位",
      tooWeak: "{label}強度不足：至少 8 位，且包含大小寫字母和數字",
      needLower: "{label}強度不足：至少包含一個小寫字母",
      needUpper: "{label}強度不足：至少包含一個大寫字母",
      needDigit: "{label}強度不足：至少包含一個數字",
      ok: "強度合格",
      levelWeak: "弱",
      levelFair: "一般",
      levelGood: "良好",
      levelStrong: "強",
    },
  },

  "en-US": {
    common: {
      permanent: "Permanent",
      notActivated: "Not activated",
      invalidDate: "Invalid date",
      relativeTime: {
        expiredAgo: "Expired {days} day(s) ago",
        today: "Expires today",
        tomorrow: "Expires tomorrow",
        inDays: "Expires in {days} day(s)",
        inWeeks: "Expires in {weeks} week(s)",
        inMonths: "Expires in {months} month(s)",
        inYears: "Expires in {years} year(s)",
      },
      httpError: {
        unauthorized: "Your session has expired, please sign in again",
        forbidden: "Insufficient permissions",
        notFound: "The requested resource was not found",
        conflict: "Operation conflict, please refresh and retry",
        payloadTooLarge: "Upload is too large",
        tooManyRequests: "Too many requests, please try again later",
        serverError: "The server ran into a problem, please try again later",
        requestFailed: "Request failed",
      },
    },
    errorBoundary: { errorId: "Error ID" },
    validators: {
      usernameRequired: "Please enter a username",
      usernameLength: "Username must be {min}-{max} characters",
      usernameFormat: "Invalid username: only letters / digits / underscores are allowed, and it cannot start with a digit",
      embyUsernameRequired: "Please enter an Emby username",
      embyUsernameTooLong: "Emby username cannot exceed {max} characters",
      embyUsernameControlChars: "Emby username contains invisible characters",
      emailTooLong: "Email cannot exceed 254 characters",
      emailFormat: "Invalid email format",
      codeRequired: "Please enter a registration / invite code",
      codeLength: "Code must be {min}-{max} characters",
      codeFormat: "Code may only contain letters / digits / underscores / hyphens",
      intMustBeInteger: "{label} must be an integer",
      intMin: "{label} cannot be less than {min}",
      intMax: "{label} cannot be greater than {max}",
    },
    password: {
      defaultLabel: "New password",
      required: "Please enter {label}",
      tooLong: "{label} is too long (max 128 characters)",
      tooWeak: "{label} is too weak: at least 8 characters including upper- and lower-case letters and a number",
      needLower: "{label} is too weak: include at least one lowercase letter",
      needUpper: "{label} is too weak: include at least one uppercase letter",
      needDigit: "{label} is too weak: include at least one number",
      ok: "Strength OK",
      levelWeak: "Weak",
      levelFair: "Fair",
      levelGood: "Good",
      levelStrong: "Strong",
    },
  },
};

// deep-merge src into dst, only adding keys that don't already exist (idempotent, non-destructive)
function mergeInto(dst, src, path = "") {
  let added = 0;
  for (const k of Object.keys(src)) {
    const p = path ? `${path}.${k}` : k;
    if (src[k] && typeof src[k] === "object" && !Array.isArray(src[k])) {
      if (!(k in dst) || typeof dst[k] !== "object") dst[k] = {};
      added += mergeInto(dst[k], src[k], p);
    } else if (!(k in dst)) {
      dst[k] = src[k];
      added++;
    } else {
      console.log("  SKIP existing:", p);
    }
  }
  return added;
}

const fileMap = {
  basic: "src/locales/basic.json",
  "zh-Hant": "src/locales/zh-Hant.json",
  "en-US": "src/locales/en-US.json",
};

for (const [name, add] of Object.entries(additions)) {
  const file = fileMap[name];
  const obj = JSON.parse(fs.readFileSync(file, "utf8"));
  const n = mergeInto(obj, add);
  fs.writeFileSync(file, JSON.stringify(obj, null, 2) + "\n", "utf8");
  console.log(`${file}: added ${n} keys`);
}
