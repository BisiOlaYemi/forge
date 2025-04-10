package mailer

import (
	"fmt"
	"html/template"
	"path/filepath"

	"gopkg.in/mail.v2"
)

// Mailer handles email sending
type Mailer struct {
	dialer    *mail.Dialer
	templates *template.Template
	from      string
}

// Config represents mailer configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TemplateDir string
}

// New creates a new mailer
func New(config Config) (*Mailer, error) {
	dialer := mail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	dialer.SSL = true

	// Test connection
	s, err := dialer.Dial()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	s.Close()

	// Load templates
	templates, err := template.ParseGlob(filepath.Join(config.TemplateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return &Mailer{
		dialer:    dialer,
		templates: templates,
		from:      config.From,
	}, nil
}

// Send sends an email
func (m *Mailer) Send(to, subject, templateName string, data interface{}) error {
	// Get template
	tmpl := m.templates.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("template %s not found", templateName)
	}

	// Render template
	var body string
	if err := tmpl.ExecuteTemplate(&body, templateName, data); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Create message
	msg := mail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	// Send message
	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendWithAttachments sends an email with attachments
func (m *Mailer) SendWithAttachments(to, subject, templateName string, data interface{}, attachments []string) error {
	// Get template
	tmpl := m.templates.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("template %s not found", templateName)
	}

	// Render template
	var body string
	if err := tmpl.ExecuteTemplate(&body, templateName, data); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Create message
	msg := mail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	// Add attachments
	for _, attachment := range attachments {
		msg.Attach(attachment)
	}

	// Send message
	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
} 