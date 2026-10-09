package usecase

import (
	"strings"

	"github.com/your-org/webapp-template/internal/domain"
)

// SignupPolicy は、誰がログインでき、誰が初回ログインで admin になるかを決める設定です。
//
// ゼロ値は「ドメインの制限なし・初期管理者なし」で、新規ユーザーは全員 member になります。
// 「最初にログインした人を admin にする」動作はありません。デプロイ直後の URL を
// 先に開いた第三者が管理者になるのを防ぐためです。
type SignupPolicy struct {
	initialAdminEmails  map[string]bool
	allowedEmailDomains map[string]bool
}

// ParseSignupPolicy は、カンマ区切りの環境変数の値から SignupPolicy を作ります。
//
//   - initialAdminEmails: 初回ログインで admin にするメールアドレス（INITIAL_ADMIN_EMAILS）
//   - allowedEmailDomains: ログインを許可するメールのドメイン（ALLOWED_EMAIL_DOMAINS）。空なら制限しない
//
// 大文字と小文字は区別しません。ドメインは先頭の @ を付けても付けなくても同じです。
func ParseSignupPolicy(initialAdminEmails, allowedEmailDomains string) SignupPolicy {
	p := SignupPolicy{
		initialAdminEmails:  make(map[string]bool),
		allowedEmailDomains: make(map[string]bool),
	}
	for _, email := range splitList(initialAdminEmails) {
		p.initialAdminEmails[email] = true
	}
	for _, d := range splitList(allowedEmailDomains) {
		if d = strings.TrimPrefix(d, "@"); d != "" {
			p.allowedEmailDomains[d] = true
		}
	}
	return p
}

// RestrictsDomains は、ログインできるメールのドメインを制限しているかどうかを返します。
func (p SignupPolicy) RestrictsDomains() bool {
	return len(p.allowedEmailDomains) > 0
}

// HasInitialAdmins は、初期管理者が 1 人以上指定されているかどうかを返します。
func (p SignupPolicy) HasInitialAdmins() bool {
	return len(p.initialAdminEmails) > 0
}

// allowsEmail は、このメールアドレスでのログインを許可するかどうかを返します。
func (p SignupPolicy) allowsEmail(email string) bool {
	if !p.RestrictsDomains() {
		return true
	}
	// 最後の @ より後ろをドメインとして扱う。サブドメインは別のドメインとして比較する
	i := strings.LastIndex(email, "@")
	if i < 0 {
		return false
	}
	return p.allowedEmailDomains[strings.ToLower(email[i+1:])]
}

// initialRole は、新規ユーザーに与えるロールを返します。
func (p SignupPolicy) initialRole(email string) domain.Role {
	if p.initialAdminEmails[strings.ToLower(strings.TrimSpace(email))] {
		return domain.RoleAdmin
	}
	return domain.RoleMember
}

// splitList はカンマ区切りの文字列を、前後の空白を除いた小文字の要素に分けます。空の要素は捨てます。
func splitList(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.ToLower(strings.TrimSpace(v)); v != "" {
			out = append(out, v)
		}
	}
	return out
}
