# 版本历史

本文记录 Twilight 各版本的变更与发布文案，按版本从新到旧排列；文末附「发布检查清单」。术语与跨文档引用见 [文档导航](../README.md)。

## 未发布（开发中）

### 邮箱验证与找回密码（新增）

- 新增邮箱验证子系统：SMTP 发信、绑定/验证邮箱、登出态邮箱找回密码、改密二次校验、强制绑定门。详见 [邮箱验证与找回密码](./features/email.md)。
- **配置**：新增 `[Email]` 段（SMTP 主机/端口/凭据/加密、`force_bind`、验证码长度/类型/有效期/重发冷却/尝试上限、邮件标题/正文模板）；邮箱域名黑白名单 `email_validation_mode` / `email_whitelist` / `email_blacklist` 归在 `[SAR]` 段，与注册校验共用。配置模板见 `config.production.toml`（功能配置统一在 `config.toml`，不走 `.env`；密钥放 `config.local.toml`），后台「系统配置 → 邮箱验证」可视化修改并热重载。
- **验证码安全**：只存服务端 HMAC-SHA256 哈希（不存明文）+ 定长常量时间比较；带尝试上限、有效期、IP/收件地址双发码限流与重发冷却；过期码由调度任务清理。HMAC 密钥优先由 `bot_internal_secret` 派生，缺省回退进程级随机密钥（重启即作废，fail-closed）。
- **强制绑定**：`force_bind` 开启时普通用户/白名单进仪表盘前须验证邮箱（管理员豁免仅提示），改系统/Emby 密码也要求邮箱验证码；服务端 `requireEmailVerified` 硬门，价值型接口不可绕过。SMTP 未配置完整时强制门自动失效，防止把用户锁在仪表盘外。
- **找回密码防枚举**：邮箱找回统一成功响应、仅命中已验证邮箱账号、吞掉内部失败；按 IP 限流。
- **管理员邮箱管理区**：用户管理页新增邮箱验证状态筛选（已验证/未验证/已填/未填）与行内状态徽标；可强制把用户绑定到指定邮箱、置/撤销验证状态；配置页内置「发送测试邮件」验证 SMTP 连通性。
- **前端与 i18n**：注册/登录忘记密码页支持邮箱验证码重置（与 Emby 找回并列），`email.*` 文案三语（basic / en-US / zh-Hant）对齐。
- **发码多维限流**：发码依次过 IP → 登录账号(uid) → 收件地址三道闸 + 重发冷却；`email_code_uid_per_10m` 防同一账号轮换收件邮箱滥刷。前端收到限流/冷却错误码会本地起冷却自禁发送按钮。

### Emby 登录用户设备 / IP 审查（新增）

- 新增 `GET /api/v1/admin/emby/devices`（`handleAdminEmbyDevices`）：以 Emby `/Devices` 设备清单为基底，用实时 `/Sessions` 的 `RemoteEndPoint` 补当前 IP 与在线状态，并把每台设备的 Emby 用户映射回本地账号；管理后台「Emby 管理 → 设备 / IP 审查」可搜索查看。实时会话 DTO 也补上了 `remote_endpoint`(IP)。详见 [安全加固 §10](./guides/security.md)。

### 注册码 / 续期码管理增强

- `PUT /api/v1/admin/regcodes/:code` 由「仅改备注」扩展为部分更新：支持停用/启用（`active`）、有效期（`validity_time` 小时，-1 永久）、授予天数、使用次数上限；DTO 新增计算字段 `expires_at`。管理页提供编辑弹窗（启用开关 + 有效期 + 额度）。详见 [注册码与卡码](./features/regcodes.md)。

### Telegram 用户名自动刷新

- Bot 处理任意来自已绑定用户的更新（私聊 / 群消息 / chat_member）时被动刷新存储的 `telegram_username`（无额外 API 调用，空用户名不清空）。详见 [Telegram Bot](./features/telegram-bot.md)。

### 本地登录设备记录修复

- Web 登录改用 `UpdateDevice` 读改写，修复此前每次登录用 `UpsertDevice` 整条覆盖、把设备的「受信任 / 已封禁」标记与首次登录时间冲掉的问题；设备记录新增 `LastIP`。新增可选设备数上限 `EnforceDeviceLimit`（受 `[DeviceLimit].device_limit_enabled` 开关，淘汰最旧的未受信任设备，绝不踢当前/受信任设备）。

### 配置方式收敛到 config.toml

- 功能配置一律放 `config.toml`（参考 `config.production.toml`，含新增 `[RateLimit]` 段），密钥放 `config.local.toml`；后端 `.env.example` 精简为仅保留监听地址 / 站点名，前端展示项（API 基址 / 站点名 / 介绍 / 图标）走 `webui/.env`。

### 管理员邀请树性能

- 邀请树渲染由「每行重复 `findRoot`/`subtreeSize`」的 O(n²) 优化为一次 O(n) 预计算 `rootOf` / `descendants` 查表，大树场景明显更流畅。

## 社交平台发布文案：Go 后端重构版

Twilight 0.0.4 Go 后端重构版已更新。

这次更新不是一次普通小修，而是把 Twilight 的生产主线正式切换到 Go 后端：更稳定的运行方式、更完整的数据迁移、更清晰的运维入口，也为后续长期维护打基础。

本次重点更新：

1. Go 后端成为主线
   - 后端入口统一为 `twilight api / all / bot / scheduler`。
   - 旧 Python 后端不再作为生产运行路径。
   - 业务模块重新拆分，Emby、TMDB、Bangumi、求片、邀请、卡码、调度、数据库和系统运维逻辑更清晰。

2. 数据库与迁移全面补齐
   - 生产模板默认推荐 PostgreSQL。
   - 保留 Go JSON 状态存储作为兼容选项。
   - 新增数据库状态、备份、恢复、迁移预检和迁移执行接口。
   - 支持从旧 SQLite 数据库只读迁移用户、卡码、邀请、公告、求片、播放记录、调度记录、Telegram 花名册等数据。
   - 恢复和迁移前会自动创建保护性备份，并要求预览和二次确认。

3. 管理后台可观测性增强
   - 新增实时日志页面。
   - 日志支持脱敏 Token、Cookie、密码、API Key、DSN 等敏感字段。
   - 新增服务器状态、Go Runtime、内存、主机负载、数据库和 Redis 状态展示。
   - 日志等级和保留行数支持配置热重载。

4. 配置与部署更适合生产
   - 配置固定使用项目运行目录下的 `config.toml`，支持 `config.local.toml` 做本地私密覆盖。
   - 支持环境变量覆盖关键配置。
   - 提供 Linux systemd 一键设置脚本。
   - 支持 API、Bot、Scheduler 分服务运行。
   - Git 自动更新接口加入 dry-run、分支校验、脏工作区保护和 fast-forward 限制。

5. 安全加固
   - 凭据型 CORS 不再接受通配符。
   - Cookie 变更类请求不再要求额外令牌；浏览器端依赖 HttpOnly 会话 Cookie，机器调用使用 Bearer Token 或 API Key。
   - 管理接口统一收紧鉴权边界。
   - 上传资源、背景样式、备份恢复和 Git 更新都加强了路径、类型和输入校验。
   - 非系统管理员绑定 Emby 管理员账号时会被限制敏感操作，防止越权。
   - 自助续期必须消耗有效注册码，不再允许无条件免费续期。
   - 新增违规审计，支持诱饵码、指名码越权记录和处罚动作。

6. 邀请、卡码和用户体验优化
   - 邀请码支持指名用户使用。
   - 邀请制用户到期后保留登录能力，便于自行续期。
   - Emby 用户上限会计入未使用的邀请码，降低超发风险。
   - 用户排序、移动端布局、配置页、注册码页、邀请树和媒体页都有优化。

7. Telegram Bot 增强
   - 账号类命令强制私聊，减少群聊泄露风险。
   - 管理员可进行只读查询。
   - 支持绑定状态、用户摘要、播放统计、服务统计和群组成员安全管理。
   - Bot 输出避免展示密码、Token、Emby ID、服务器线路等敏感信息。

8. CI 和运行环境升级
   - GitHub Actions 已迁移为 Go 测试、安全扫描和 Nix 检查。
   - Go 版本升级到 1.25+，规避旧标准库安全漏洞。
   - 后端通过 `go test ./...`、`go vet ./...` 和 govulncheck 检查。

适合谁升级：

- 正在使用 Twilight 管理 Emby / Jellyfin 用户的站点。
- 想从旧 Python 后端迁移到 Go 后端的用户。
- 需要 PostgreSQL、运行日志、数据库迁移、系统更新和更完整后台运维能力的管理员。

升级提醒：

- 升级前请备份配置、数据库和上传目录。
- 旧 SQLite 用户请先阅读迁移文档，使用后台数据库迁移页预检后再执行。
- 生产环境建议使用 PostgreSQL、HTTPS、明确 CORS Origin，并妥善保管 `config.local.toml` 和密钥文件。

项目地址：<https://github.com/Prejudice-Studio/Twilight>
频道：<https://t.me/Twilightpanel>
交流群：<https://t.me/TwilightPanelChat>

## 0.0.4 - 2026-05-23（当前）

本版本聚焦 Go 后端安全加固、数据库迁移补齐、日志可观测性、前后端接口对齐和管理员体验优化。

### 运行日志与状态面板

- `[Global]` 新增 `log_level` 与 `runtime_log_limit`，兼容旧 Python 数字日志等级 `10/20/30/40`，配置保存后可热重载。
- 管理后台实时日志支持查看更多历史行数，后端按配置保留脱敏日志缓冲。
- `/api/v1/system/health` 对齐前端，返回 `api/database/emby` 健康字段。
- `/api/v1/system/admin/stats` 补齐 Emby、注册码、用户统计结构，服务器状态面板可正常显示。

### 数据库迁移与兼容

- 数据库状态接口明确标注 `gojson`、`sqlite3`、`postgresql` 类型和用途。
- 前端可视化配置默认推荐 PostgreSQL；SQLite 保留为手动迁移来源，不在前端作为运行后端选项展示。
- 旧 SQLite 迁移补齐 Telegram 用户名、Emby 用户名、求片 UID/TG/状态/备注/更新时间映射，降低迁移后绑定状态与求片状态不同步的问题。

### 公告与邀请树

- 公告模型补齐 `render_mode`、`pinned`、`expires_at`、`created_by_uid`，Markdown / BBCode 选择会被后端持久化。公告以字段形式保存在单一状态文档（`internal/store` 的 `Announcements` 映射）中，无需建表或迁移。
- 公告渲染器清理控制字符转义，继续保持不使用 `dangerouslySetInnerHTML` 的安全渲染策略。
- 管理员邀请树星图增加鼠标动态光照、邻近节点与边高亮，保持大树场景下的轻量渲染。

### 用户管理与安全

- 用户列表排序补齐 `uid_asc/uid_desc/username_desc/register_time_asc/expired_at_desc` 等前端选项，避免排序与分页结果不一致。
- 非系统管理员一旦绑定 Emby 管理员账号，除查看账号状态和退出登录外，所有已鉴权业务请求都会被统一拒绝。
- Emby 管理员身份检查增加短缓存，降低频繁访问 Emby API 对性能的影响。

### 安全加固（关键）

- **Emby 管理员账号隔离**：非系统管理员用户绑定了 Emby 管理员账号时，禁止所有敏感操作（修改密码、修改 Emby 密码、修改个人资料、解绑 Emby 等），防止越权控制 Emby 服务器。
- **绑定时拦截**：非系统管理员用户不允许绑定 Emby 管理员账号，从源头阻断风险。
- **自助续期修复**：`/users/me/renew` 不再允许无条件免费续期，必须提供有效注册码，通过 `ConsumeRegCode` 验证并消耗。
- **Telegram 强制绑定策略**：`handleUnbindTelegram` 现在正确检查 `force_bind_telegram` 配置，非管理员用户在强制绑定模式下无法解绑。`handleTelegramStatus` 返回真实的 `force_bind` 和 `can_unbind` 状态。
- **管理员保护**：`handleAdminSetRole` 新增最后管理员保护，禁止移除系统中唯一活跃管理员的权限。
- **输入验证加强**：注册、修改用户名、修改个人资料接口增加 `<>"'&` 等 XSS 危险字符过滤；邮箱增加长度和字符校验。
- **路由鉴权修正**：`PUT /api/v1/media/request/:request_id/status` 从 `AuthUser` 提升为 `AuthAdmin`，与 handler 内部检查一致。
- **Emby 线路越权修复**：`handleEmbyURLs` 新增对已禁用账号（`!u.Active`）的检查，防止 Emby 被禁用或过期后仍能查看服务器线路。

### 邀请制度完善

- **邀请码指名限制**：生成邀请码时可指定 `target_username`，只有匹配的用户名才能使用该邀请码。
- **邀请制用户到期策略**：通过邀请注册的用户到期时只禁用 Emby 访问，保持账号 Active，用户仍可登录续期。非邀请制用户行为不变。
- **Emby 用户上限计算优化**：容量检查现在额外计入活跃未使用的邀请码（每个代表一个潜在 Emby 名额），防止超发。

### 违规审计系统（新增）

- **诱饵码检测与处罚**：使用诱饵注册码时自动记录违规并执行配置的处罚动作（`regcode_decoy_action`：`disable_user` / `disable_emby` / `log_only`）。
- **指名码越权检测**：注册码设置 `target_username` 后，非目标用户尝试使用时记录违规并拒绝。
- **违规审计接口**：`GET /api/v1/admin/violations`（分页、筛选）、`DELETE /api/v1/admin/violations/:id`、`POST /api/v1/admin/violations/clear`。
- **前端审计页面**：新增 `/admin/violations` 管理页面，支持按类型筛选、搜索、单条删除和全部清除。
- **Store 层**：新增 `ViolationLog` 结构体和 `AddViolationLog`、`ListViolationLogs`、`DeleteViolationLog`、`ClearViolationLogs` 方法。
- **数据库兼容**：新增字段均使用 `omitempty` 或 nil-safe slice，以字段形式存在于单一状态文档中，旧备份恢复到新版本无需迁移。

### 服务器图标

- 服务器图标从内联 SVG 改为嵌入式 PNG（使用项目 Logo），通过 `go:embed` 编译进二进制。
- `/api/v1/system/server-icon` 返回 `image/png` 并设置 `Cache-Control: public, max-age=86400`。

### CI/CD

- GitHub Actions 工作流从旧 Python（pytest/mypy/bandit）完全迁移到 Go。
- 新增 `go-test` job：多平台（ubuntu/windows）运行 `go vet` + `go test -race`。
- 新增 `go-security` job：运行 `govulncheck` 检查已知漏洞。
- 保留 `nix` job：验证 flake 输出和 dev shell 可用性。

### 代码质量

- 清理重复的过程性注释，精简函数文档。
- 前端 `utils.ts` 改为函数级 JSDoc。
- 所有 Go 测试通过（`go test ./...`），`go vet` 无警告。

## 0.0.4 - 2026-05-22（Go 后端重构基线）

本版本将 `golang` 分支确认为 Go 后端主线版本，目标是完整承接旧后端能力、对齐前端接口，并补齐部署、数据库、安全和移动端体验。

### 后端架构

- 将后端业务按功能域拆分，避免继续把 Emby、TMDB、Bangumi、求片、邀请、卡码、调度、数据库和系统更新逻辑混在少数大文件中。
- `internal/api` 现在按 handler、client、service 和运维模块维护，外部服务调用统一收敛到独立客户端文件。
- `cmd/twilight` 统一提供 `api`、`all`、`scheduler`、`bot`、`version` 子命令；Telegram 未启用时 bot 子命令安全空转，避免 systemd 重启循环。

### 数据库与迁移

- `config.production.toml` 新增 `[Database]` PostgreSQL 配置示例，包含完整 DSN、分项连接参数、备份目录和连接池参数。
- 新增数据库状态、备份、恢复、迁移预检和迁移执行接口，并提供独立的数据库迁移管理页。
- 默认数据库后端切换为 PostgreSQL；Go JSON 状态存储保留为兼容选项，迁移预检会返回实体数量、快照大小、目标连通性和配置/重启告警。
- PostgreSQL 目标数据库不存在时，启动阶段和迁移预检会尝试连接 `postgres` / `template1` 维护库自动执行 `CREATE DATABASE`，减少已有用户但未建库时的部署阻塞。
- 当配置为 PostgreSQL 但目标库尚未迁移且没有管理员时，启动阶段会检测旧 JSON 状态文件；若其中已有 active 管理员，则临时回退 JSON，保证原管理员可以登录管理端执行迁移。
- 当 Go 状态没有 active 管理员但存在旧 Python 版 `db/users.db` 时，启动阶段可通过系统 `sqlite3` 只读导入旧库 active 管理员账号用于引导登录。
- 数据库管理页会自动检测多份旧 SQLite 数据库；备份会同时复制 `.db`、`.db-wal`、`.db-shm` 文件集。
- 迁移来源新增「旧 SQLite」，按固定库名映射用户、API Key、注册码、邀请码、公告、Bangumi/TMDB 求片、签到积分、登录设备、播放记录、调度记录和 Telegram 花名册，再写入 JSON 或 PostgreSQL。
- PostgreSQL 迁移预检会准备目标数据库和 `twilight_state` 状态表，但不会写入业务快照；目标库不存在且用户有 `CREATEDB` 权限时可自动创建。
- 恢复和迁移执行前强制预览与二次确认；后端在缺少确认短语时只返回 `dry_run=true` 预览，不执行写入。
- 恢复备份和执行迁移前都会自动创建保护性备份，并在响应中返回 `pre_operation_backup` 方便审计和回滚。
- 备份恢复限制在备份目录内，只接受普通 `.json` 文件并拒绝符号链接。
- JSON 迁移目标限制在数据库目录内，且必须是 `.json` 文件，防止路径穿越和错误文件类型写入。

### 安全加固

- 凭据型 CORS 不再接受 `*`，生产环境必须显式配置可信前端 Origin。
- CORS Origin 匹配增加规范化处理，允许配置尾斜杠，但拒绝带路径、查询串、片段或非法 scheme 的来源，降低跨域配置误用风险。
- Cookie 变更类请求不再要求额外令牌；Cookie 写请求只依赖有效会话，Bearer Token 与 API Key 按各自鉴权路径处理，`X-Twilight-Client` 不参与鉴权。
- 新增管理员实时日志与服务器状态接口；日志只来自 Go 进程内缓冲，不读取任意系统日志文件，并对 Token、Cookie、密码、API Key、DSN 等敏感内容做脱敏。
- 用户背景配置改为后端结构化校验，只允许安全渐变表达式和本系统上传的背景资源，阻断任意外部 URL、复杂 CSS 函数和 `url()` 注入。
- 标准前端部署禁用 Next 服务端图片优化器，避免特殊部署环境中由图片优化器代拉任意远程 URL。
- 上传入口只允许 JPEG、PNG、GIF、WebP、BMP 这类栅格图片 MIME；读取上传资源时只接受服务端生成的白名单文件名。
- 上传资源读取路径会重新计算绝对路径并校验仍位于上传目录内，防止路径穿越、转义路径和伪造文件名。
- Git 更新接口仅接受不含凭据的 HTTPS 仓库 URL，校验分支名，默认拒绝脏工作区，执行前支持 `dry_run` 预检，并使用 `git pull --ff-only`；只有 commit 变化时才调度重启，优先通过 `systemd-run` 延迟重启服务。
- Telegram Bot 账号类命令强制私聊，群聊仅保留管理员只读查询；绑定码增加格式校验，Bot 输出不展示密码、Token、Emby ID、服务器线路等敏感信息。

### Linux 部署

- 新增 `deploy/setup-systemd.sh`，用于 Linux 一键设置 systemd 服务。
- 脚本会检查项目路径、配置路径、二进制路径、服务用户/组、端口、依赖命令、systemd 特殊字符和旧 Python 版 systemd 残留。
- 检测到旧 Python unit 时会停止、禁用并备份旧服务文件，再写入 Go 版 unit。
- systemd unit 默认指向 `bin/twilight`，并使用独立的 API、Bot、Scheduler 服务名保持向后兼容。

### 前端体验

- 新增管理员「实时日志」页面，可实时查看后端日志流、Go 运行时、主机负载、内存、数据库和 Redis 状态。
- 优化移动端主布局、顶部栏、导航抽屉、配置页、注册码页、媒体页和邀请树视图。
- 注册码页在移动端切换为卡片列表，桌面端保留高密度表格。
- 配置页移动端标签改为两列网格，避免内部横向滚动。
- 邀请树加入更适合移动端的缩放、平移和统计布局。

### 文档与维护

- 文档统一以 Go 后端和 Linux 部署路径为准，移除旧平台快速启动入口。
- 更新 Go 后端说明、安装部署、开发维护、API Key 示例、版本历史和发布检查清单。
- `.gitignore` 扩展覆盖 Go 构建产物、前端缓存、运行数据、数据库文件、本地密钥、补丁备份和 1Panel 本地运行配置。

## 0.0.3

### 媒体与外部集成

- 新增 Bangumi 同步流程，并对齐前端的同步状态显示。

### 邀请与卡码

- 支持更细粒度的邀请码行为配置。
- 新增使用码预览能力，用户在真正消费注册码、续期码、白名单码或邀请码前可以先看到效果。

### 前端兼容

- 优化前端 API 错误处理，后端能力缺失或接口不匹配时给出更清晰的用户提示。

### 安全与稳定性

- 加强接口限流和错误响应一致性。
- 加强 Emby 密码验证、会话检查和容量检查，降低误注册、超额使用和异常会话风险。

## 0.0.2

### 安全基础

- 增加面向公网部署的安全修复和 Emby 容量控制。
- 增加代理感知的客户端 IP 识别，提升反向代理部署下的限流准确性。
- 增加 Redis 限流和会话能力，支持多实例共享状态。
- 增加公网部署安全护栏。

### 数据一致性

- 增加 SQLite pragma 和事务化注册流程，降低并发注册和异常退出导致的数据损坏风险。
- 增加注册队列、过期清理、待补建 Emby 用户处理和注册权益补发工具。

### 邀请与注册码

- 增加根用户邀请限制和更严格的注册准入控制。
- 增加注册码格式、随机算法、诱饵码隐藏、注册码用户查询和使用队列。
- 支持通过注册码授予 Emby 注册权益、Emby 重置、过期处理和取消永久有效状态。

### 系统更新

- 增加 Git 自动更新 API 和管理端入口。
- 增加系统自动更新和注册队列运维工具。

### 文档

- 补充工作流、版本操作和运维说明。

## 0.0.1

### 项目基础

- 建立 Twilight 后端和前端基础能力，包括注册、登录、角色检查、管理员用户管理和 Emby 绑定/注册。
- 建立 dashboard、设置、用户管理、管理员管理、注册码、邀请、Emby 操作、公告和服务信息等基础页面。

### Telegram

- 增加 Telegram Bot 集成、自定义 Bot 消息、群组管理员工具和相关 API 文档。

### 邀请与卡码

- 增加邀请码、注册码、续期码、白名单码和早期管理员运维流程。

### 文档

- 增加中文项目说明和生产启动指导。

## 发布检查清单

- 更新后端版本号：`cmd/twilight/main.go`、`internal/config/config.go`。
- 更新前端版本号：`webui/package.json`、`webui/package-lock.json`。
- 对变更的 Go 文件执行 `gofmt`。
- 执行 `go test ./...`。
- 执行 `go vet ./...`。
- 修改前端或 API 客户端后，在 `webui/` 执行 lint 和 build。
- 发布前扫描敏感信息、旧后端残留、路径穿越、文件类型白名单和鉴权边界。
