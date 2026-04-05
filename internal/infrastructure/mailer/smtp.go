package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/domain"
)

type smtpMailer struct {
	host string
	port int
	from string
	auth smtp.Auth // nil when no credentials are configured (e.g. Mailhog)
	tls  bool      // true only on port 465 (implicit TLS)
}

func New(cfg *config.MailerConfig) (domain.Mailer, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("mailer: MAILER_HOST is required")
	}

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return &smtpMailer{
		host: cfg.Host,
		port: cfg.Port,
		from: cfg.From,
		auth: auth,
		tls:  cfg.Port == 465,
	}, nil
}

func (m *smtpMailer) Send(_ context.Context, msg domain.MailMessage) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("mailer: at least one recipient is required")
	}

	raw := m.buildRaw(msg)
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	if m.tls {
		return m.dialTLS(addr, msg.To, raw)
	}

	// Works for both Mailhog (no auth) and STARTTLS providers (with auth).
	return smtp.SendMail(addr, m.auth, m.from, msg.To, raw)
}

// dialTLS is used only for implicit TLS on port 465.
func (m *smtpMailer) dialTLS(addr string, to []string, raw []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.host})
	if err != nil {
		return fmt.Errorf("mailer: tls dial: %w", err)
	}

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("mailer: smtp client: %w", err)
	}
	defer client.Close()

	if m.auth != nil {
		if err = client.Auth(m.auth); err != nil {
			return fmt.Errorf("mailer: smtp auth: %w", err)
		}
	}

	if err = client.Mail(m.from); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("mailer: RCPT TO %s: %w", rcpt, err)
		}
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err = wc.Write(raw); err != nil {
		return fmt.Errorf("mailer: write body: %w", err)
	}
	return wc.Close()
}

// buildRaw assembles a multipart/alternative MIME message (HTML + plain-text).
func (m *smtpMailer) buildRaw(msg domain.MailMessage) []byte {
	const boundary = "==boundary_studio=="

	plain := msg.Text
	if plain == "" {
		plain = stripTags(msg.HTML)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", m.from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(msg.To, ", "))
	if len(msg.CC) > 0 {
		fmt.Fprintf(&b, "Cc: %s\r\n", strings.Join(msg.CC, ", "))
	}
	if msg.ReplyTo != "" {
		fmt.Fprintf(&b, "Reply-To: %s\r\n", msg.ReplyTo)
	}
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)

	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", boundary, plain)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n", boundary, msg.HTML)
	fmt.Fprintf(&b, "--%s--\r\n", boundary)

	return []byte(b.String())
}

func stripTags(html string) string {
	var out strings.Builder
	in := false
	for _, r := range html {
		switch {
		case r == '<':
			in = true
		case r == '>':
			in = false
		case !in:
			out.WriteRune(r)
		}
	}
	return strings.TrimSpace(out.String())
}
