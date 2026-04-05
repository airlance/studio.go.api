package domain

import "context"

// MailMessage represents a single outgoing email.
// All fields are plain Go types — no transport-specific dependencies.
type MailMessage struct {
	To      []string
	Subject string
	HTML    string   // rendered HTML body
	Text    string   // optional plain-text fallback
	ReplyTo string   // optional reply-to address
	CC      []string // optional CC recipients
}

// Mailer is the port that the service layer uses to dispatch emails.
// Implementations live in internal/infrastructure/mailer/.
type Mailer interface {
	Send(ctx context.Context, msg MailMessage) error
}
