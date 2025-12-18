package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"strings"

	"github.com/jordan-wright/email"
)

// SMTPService implements Service interface using SMTP
type SMTPService struct {
	config      *SMTPConfig
	defaultFrom string
}

// NewSMTPService creates a new SMTP email service
func NewSMTPService(config *SMTPConfig, defaultFrom string) (*SMTPService, error) {
	if config == nil {
		return nil, fmt.Errorf("SMTP configuration is required")
	}

	if config.Host == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}

	if config.Port == 0 {
		config.Port = 587 // Default SMTP port
	}

	return &SMTPService{
		config:      config,
		defaultFrom: defaultFrom,
	}, nil
}

// Send sends an email
func (s *SMTPService) Send(ctx context.Context, msg *Email) error {
	e := email.NewEmail()

	// Set from address
	if msg.From != "" {
		e.From = msg.From
	} else if s.defaultFrom != "" {
		e.From = s.defaultFrom
	} else {
		return fmt.Errorf("from address is required")
	}

	// Set recipients
	if len(msg.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	e.To = msg.To
	e.Cc = msg.CC
	e.Bcc = msg.BCC

	// Set subject
	e.Subject = msg.Subject

	// Set body
	if msg.Body != "" {
		e.Text = []byte(msg.Body)
	}
	if msg.HTML != "" {
		e.HTML = []byte(msg.HTML)
	}

	// Add attachments
	for _, att := range msg.Attachments {
		if att.Data != nil {
			e.Attach(strings.NewReader(string(att.Data)), att.Filename, att.ContentType)
		} else if att.Reader != nil {
			data, err := io.ReadAll(att.Reader)
			if err != nil {
				return fmt.Errorf("failed to read attachment: %w", err)
			}
			e.Attach(strings.NewReader(string(data)), att.Filename, att.ContentType)
		}
	}

	// Add custom headers
	for key, value := range msg.Headers {
		e.Headers.Add(key, value)
	}

	// Send email
	var auth smtp.Auth
	if s.config.Username != "" && s.config.Password != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	if s.config.TLS {
		// Use TLS
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}
		return e.SendWithTLS(addr, auth, tlsConfig)
	}

	// Use STARTTLS
	return e.Send(addr, auth)
}

// SendTemplate sends a template-based email
func (s *SMTPService) SendTemplate(ctx context.Context, opts *TemplateEmailOptions) error {
	// Note: Template rendering should be done by the caller
	// This is a placeholder implementation
	return fmt.Errorf("template email not implemented in SMTP service - render template first")
}
