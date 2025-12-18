package email

import (
	"context"
	"io"
)

// Service defines the interface for email operations
type Service interface {
	Send(ctx context.Context, email *Email) error
	SendTemplate(ctx context.Context, opts *TemplateEmailOptions) error
}

// Email represents an email message
type Email struct {
	From        string        `json:"from"`
	To          []string      `json:"to"`
	CC          []string      `json:"cc,omitempty"`
	BCC         []string      `json:"bcc,omitempty"`
	Subject     string        `json:"subject"`
	Body        string        `json:"body,omitempty"`        // Plain text body
	HTML        string        `json:"html,omitempty"`        // HTML body
	Attachments []*Attachment `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Data        []byte    `json:"-"` // Binary data
	Reader      io.Reader `json:"-"` // Alternative to Data
}

// TemplateEmailOptions contains options for template-based emails
type TemplateEmailOptions struct {
	Template    string                 `json:"template"`
	Data        interface{}            `json:"data"`
	To          []string               `json:"to"`
	CC          []string               `json:"cc,omitempty"`
	BCC         []string               `json:"bcc,omitempty"`
	Subject     string                 `json:"subject"`
	Attachments []*Attachment          `json:"attachments,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
}

// Config represents email service configuration
type Config struct {
	Provider   string      `yaml:"provider"`    // smtp, sendgrid, ses
	From       string      `yaml:"from"`        // Default from address
	SMTP       *SMTPConfig `yaml:"smtp,omitempty"`
	SendGrid   *SendGridConfig `yaml:"sendgrid,omitempty"`
	SES        *SESConfig  `yaml:"ses,omitempty"`
}

// SMTPConfig contains SMTP configuration
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	TLS      bool   `yaml:"tls"`
}

// SendGridConfig contains SendGrid configuration
type SendGridConfig struct {
	APIKey string `yaml:"api_key"`
}

// SESConfig contains AWS SES configuration
type SESConfig struct {
	Region    string `yaml:"region"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

// NewService creates a new email service based on configuration
func NewService(config *Config) (Service, error) {
	switch config.Provider {
	case "smtp", "":
		return NewSMTPService(config.SMTP, config.From)
	case "sendgrid":
		return NewSendGridService(config.SendGrid, config.From)
	case "ses":
		return NewSESService(config.SES, config.From)
	default:
		return NewSMTPService(config.SMTP, config.From)
	}
}
