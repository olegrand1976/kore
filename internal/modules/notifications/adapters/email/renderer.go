package email

import (
	_ "embed"
	"strings"
)

//go:embed templates/base.html
var baseTemplate string

//go:embed templates/notification.html
var notificationTemplate string

func RenderNotification(body, company, url string) string {
	return renderTemplate(baseTemplate, body, company, url, "")
}

func RenderTransactional(subject, body, company, url string) string {
	return renderTemplate(notificationTemplate, body, company, url, subject)
}

func renderTemplate(tpl, body, company, url, subject string) string {
	safeBody := strings.ReplaceAll(htmlEscape(body), "\n", "<br/>")
	out := strings.ReplaceAll(tpl, "{{BODY}}", safeBody)
	out = strings.ReplaceAll(out, "{{COMPANY}}", htmlEscape(company))
	out = strings.ReplaceAll(out, "{{URL}}", htmlEscape(url))
	out = strings.ReplaceAll(out, "{{SUBJECT}}", htmlEscape(subject))
	return out
}

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;").Replace(s)
}
