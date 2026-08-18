package governance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"regexp"

	"awesomeProject4/agent/knowledge/ragkit/ragkitdb"
	"gorm.io/gorm"
)

// AuditLogger 抽象审计写入。
type AuditLogger interface {
	Log(ctx context.Context, e AuditEvent) error
}

type AuditEvent struct {
	Operator       string
	Action         string
	ResourceType   string
	ResourceID     string
	Before, After  any
	Result, Reason string
	IP             string
}

type DefaultAuditLogger struct {
	db *gorm.DB
}

func NewDefaultAuditLogger(db *gorm.DB) AuditLogger { return &DefaultAuditLogger{db: db} }

func (l *DefaultAuditLogger) Log(ctx context.Context, e AuditEvent) error {
	row := ragkitdb.AuditEventRow{
		AuditTraceID:    newTraceID(),
		Operator:        e.Operator,
		Action:          e.Action,
		ResourceType:    e.ResourceType,
		ResourceID:      e.ResourceID,
		Before:          maskJSON(e.Before),
		After:           maskJSON(e.After),
		Result:          e.Result,
		Reason:          e.Reason,
		IP:              maskIP(e.IP),
		SensitiveMasked: true,
	}
	return l.db.WithContext(ctx).Create(&row).Error
}

func newTraceID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var querySnippetRe = regexp.MustCompile(`(?i)(query|snippet|content)["']?\s*[:=]\s*["'][^"']*["']`)

// maskJSON 把敏感字段（query/snippet/content）脱敏。
func maskJSON(v any) string {
	if v == nil {
		return ""
	}
	s := mustJSON(v)
	s = querySnippetRe.ReplaceAllString(s, `"$1":"[masked]"`)
	return s
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// maskIP 把 IP 第 3、4 段打码。
func maskIP(ip string) string {
	// 简化：保留前两段
	parts := splitOnDot(ip)
	if len(parts) <= 2 {
		return ip
	}
	return parts[0] + "." + parts[1] + ".*.*"
}

func splitOnDot(s string) []string {
	out := []string{}
	cur := ""
	for _, r := range s {
		if r == '.' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
