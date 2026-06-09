package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/prejudice-studio/twilight/internal/config"
	"github.com/wneessen/go-mail"
)

// errEmailNotConfigured 表示 SMTP 关键参数缺失（host / 发件地址）。调用方据此
// 回 EMAIL_DISABLED，而不是把底层错误透给前端。
var errEmailNotConfigured = errors.New("email: smtp not configured")

// emailConfigured 判定邮箱子系统是否可用：总开关开启 + 有 SMTP 主机 + 有可用的
// 发件地址（from_address 或回退 smtp_username）。前端入口展示、后端发码守卫都以
// 此为准，避免 force_bind=true 却没配 SMTP 把用户锁死在仪表盘外。
func emailConfigured(cfg *config.Config) bool {
	if cfg == nil || !cfg.EmailEnabled {
		return false
	}
	if strings.TrimSpace(cfg.SMTPHost) == "" {
		return false
	}
	return strings.TrimSpace(cfg.SMTPFromAddress) != "" || strings.TrimSpace(cfg.SMTPUsername) != ""
}

// smtpDeliver 用 go-mail 按加密方式投递一封纯文本邮件。所有参数显式传入，不依赖
// 全局配置；错误经 redactSensitiveText 脱敏后返回，避免把 SMTP 凭据 / 连接串泄露
// 到上层响应或日志。
func smtpDeliver(ctx context.Context, cfg config.Config, to, subject, body string) error {
	host := strings.TrimSpace(cfg.SMTPHost)
	if host == "" {
		return errEmailNotConfigured
	}
	from := strings.TrimSpace(cfg.SMTPFromAddress)
	if from == "" {
		from = strings.TrimSpace(cfg.SMTPUsername)
	}
	if from == "" {
		return errEmailNotConfigured
	}
	port := cfg.SMTPPort
	if port <= 0 {
		port = 587
	}
	timeout := time.Duration(cfg.SMTPTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	opts := []mail.Option{mail.WithPort(port), mail.WithTimeout(timeout)}
	// 加密方式与端口解耦：ssl=隐式 TLS（整条连接即 TLS，典型 465）；starttls=强制
	// 显式升级（典型 587）；none=明文（典型 25）。WithSSL 与 STARTTLS 策略互斥，
	// 因此分支里只设其中一个。
	switch strings.ToLower(strings.TrimSpace(cfg.SMTPEncryption)) {
	case "ssl", "tls":
		opts = append(opts, mail.WithSSL())
	case "none", "plain":
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	default: // starttls（含空值默认）
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}
	// 仅在配置了用户名时认证；AutoDiscover 让 go-mail 按服务器 EHLO 通告自动挑选
	// LOGIN / PLAIN / CRAM-MD5 等机制，省去额外配置。
	if user := strings.TrimSpace(cfg.SMTPUsername); user != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(user),
			mail.WithPassword(cfg.SMTPPassword),
		)
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		return fmt.Errorf("email client init failed: %s", redactSensitiveText(err.Error()))
	}

	msg := mail.NewMsg()
	fromName := strings.TrimSpace(cfg.SMTPFromName)
	if fromName == "" {
		fromName = strings.TrimSpace(cfg.AppName)
	}
	if fromName != "" {
		if err := msg.FromFormat(fromName, from); err != nil {
			return fmt.Errorf("invalid sender: %s", redactSensitiveText(err.Error()))
		}
	} else if err := msg.From(from); err != nil {
		return fmt.Errorf("invalid sender: %s", redactSensitiveText(err.Error()))
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("invalid recipient: %s", redactSensitiveText(err.Error()))
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	sendCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := client.DialAndSendWithContext(sendCtx, msg); err != nil {
		return fmt.Errorf("email send failed: %s", redactSensitiveText(err.Error()))
	}
	return nil
}
