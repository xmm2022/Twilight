package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// DefaultPBKDF2Iterations 是 PBKDF2-SHA256 当前的目标迭代数。
//
// 历史值为 100k（OWASP 2017 推荐）。OWASP 2024 已将下限提至 600k，
// 与最新硬件能力匹配（普通服务端单核验证耗时 ~150ms，仍在用户体验
// 可接受范围）。
//
// 旧哈希仍被 VerifyPassword 兼容，并通过 NeedsRehash 在登录时透明迁移。
const DefaultPBKDF2Iterations = 600000

// MinAcceptablePBKDF2Iterations 是接受现存哈希的最低迭代数门槛。
// 低于该值视为陈旧哈希，登录成功后应触发 rehash。
//
// 注意 VerifyPassword 仍允许 10k-1M 范围内的迭代数完成校验，避免
// 历史用户全部立即失败；NeedsRehash 才是触发升级的窗口。
const MinAcceptablePBKDF2Iterations = 600000

func RandomHex(bytesLen int) (string, error) {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func HashPassword(password string) (string, error) {
	salt, err := randomAlphaNum(16)
	if err != nil {
		return "", err
	}
	return hashPasswordWithSalt(password, salt, DefaultPBKDF2Iterations), nil
}

func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) == 2 {
		// Legacy Python format: salt$sha256(salt + password).
		salt := parts[0]
		sum := sha256.Sum256([]byte(salt + password))
		expected := salt + "$" + hex.EncodeToString(sum[:])
		return subtle.ConstantTimeCompare([]byte(expected), []byte(encoded)) == 1
	}
	if len(parts) != 3 {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations < 10000 || iterations > 1000000 {
		return false
	}
	expected := hashPasswordWithSalt(password, parts[0], iterations)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(encoded)) == 1
}

// NeedsRehash 在以下情况返回 true：
//   - 哈希格式为 legacy（"salt$sha256"）。
//   - 当前哈希迭代数低于 MinAcceptablePBKDF2Iterations。
//
// 调用方（auth_handlers.go 登录路径）应在 VerifyPassword 通过后检查此函数，
// 若需要重哈希则透明替换 user.PasswordHash。这样不会让任何用户因升级而被
// 强制重置密码，但会让所有人在下次登录时自动迁移到新参数。
func NeedsRehash(encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) == 2 {
		// legacy python sha256 ⇒ 必须升级
		return true
	}
	if len(parts) != 3 {
		// 损坏的哈希；上层 VerifyPassword 会拒绝，但这里也算"需要重新生成"
		return true
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return true
	}
	return iterations < MinAcceptablePBKDF2Iterations
}

func hashPasswordWithSalt(password, salt string, iterations int) string {
	dk := pbkdf2SHA256([]byte(password), []byte(salt), iterations, 32)
	return fmt.Sprintf("%s$%d$%s", salt, iterations, hex.EncodeToString(dk))
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hLen := sha256.Size
	numBlocks := (keyLen + hLen - 1) / hLen
	out := make([]byte, 0, numBlocks*hLen)
	for block := 1; block <= numBlocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func randomAlphaNum(length int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var b strings.Builder
	b.Grow(length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[n.Int64()])
	}
	return b.String(), nil
}
