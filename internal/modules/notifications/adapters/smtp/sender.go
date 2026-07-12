package smtp

import (
	"bytes"
	"context"
	"fmt"
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
	headers := []string{
		"From: " + s.from,
		"To: " + strings.Join(msg.To, ", "),
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

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	return smtp.SendMail(addr, nil, s.from, msg.To, body.Bytes())
}

var _ ports.EmailSender = (*Sender)(nil)
