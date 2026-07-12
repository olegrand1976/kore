package smtp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"strings"

	"github.com/kore/kore/internal/modules/notifications/ports"
)

type Sender struct {
	host string
	port int
	from string
}

func NewSender(host string, port int, from string) *Sender {
	return &Sender{host: host, port: port, from: from}
}

func (s *Sender) Send(ctx context.Context, msg ports.Email) error {
	_ = ctx
	if len(msg.To) == 0 {
		return fmt.Errorf("no recipients")
	}

	var body bytes.Buffer
	toHeader := strings.Join(msg.To, ", ")
	if msg.HTMLBody != "" {
		boundary := "kore-" + fmt.Sprintf("%d", len(msg.Body)+len(msg.HTMLBody))
		body.WriteString("From: " + s.from + "\r\n")
		body.WriteString("To: " + toHeader + "\r\n")
		body.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", msg.Subject) + "\r\n")
		body.WriteString("MIME-Version: 1.0\r\n")
		body.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n")

		writePart := func(contentType, payload string) {
			body.WriteString("--" + boundary + "\r\n")
			body.WriteString("Content-Type: " + contentType + "\r\n")
			body.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
			body.WriteString(base64.StdEncoding.EncodeToString([]byte(payload)))
			body.WriteString("\r\n")
		}
		writePart("text/plain; charset=UTF-8", msg.Body)
		writePart("text/html; charset=UTF-8", msg.HTMLBody)
		body.WriteString("--" + boundary + "--\r\n")
	} else {
		headers := []string{
			"From: " + s.from,
			"To: " + toHeader,
			"Subject: " + msg.Subject,
			"MIME-Version: 1.0",
			"Content-Type: text/plain; charset=UTF-8",
		}
		for _, h := range headers {
			body.WriteString(h)
			body.WriteString("\r\n")
		}
		body.WriteString("\r\n")
		body.WriteString(msg.Body)
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	return smtp.SendMail(addr, nil, s.from, msg.To, body.Bytes())
}

var _ ports.EmailSender = (*Sender)(nil)
