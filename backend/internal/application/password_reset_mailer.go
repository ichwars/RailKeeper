package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

type PasswordResetMailer interface {
	SendPasswordReset(ctx context.Context, toEmail, resetURL, expiresAt string) error
}

type SMTPPasswordResetMailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	TLSMode  string
}

type SMTPPasswordResetMailer struct {
	config SMTPPasswordResetMailConfig
}

func NewSMTPPasswordResetMailer(config SMTPPasswordResetMailConfig) (*SMTPPasswordResetMailer, error) {
	config.Host = strings.TrimSpace(config.Host)
	config.Port = strings.TrimSpace(config.Port)
	config.Username = strings.TrimSpace(config.Username)
	config.From = strings.TrimSpace(config.From)
	config.TLSMode = strings.ToLower(strings.TrimSpace(config.TLSMode))
	if config.Port == "" {
		config.Port = "587"
	}
	if config.TLSMode == "" {
		config.TLSMode = "starttls"
	}
	if config.Host == "" || config.From == "" {
		return nil, nil
	}
	if _, err := mail.ParseAddress(config.From); err != nil {
		return nil, fmt.Errorf("parse smtp from address: %w", err)
	}
	switch config.TLSMode {
	case "starttls", "implicit", "none":
	default:
		return nil, fmt.Errorf("unsupported smtp tls mode %q", config.TLSMode)
	}
	return &SMTPPasswordResetMailer{config: config}, nil
}

func (m *SMTPPasswordResetMailer) SendPasswordReset(ctx context.Context, toEmail, resetURL, expiresAt string) error {
	if m == nil {
		return nil
	}
	toEmail = strings.TrimSpace(toEmail)
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return fmt.Errorf("parse reset recipient: %w", err)
	}

	message := buildPasswordResetMessage(m.config.From, toEmail, resetURL, expiresAt)
	return sendSMTP(ctx, m.config, toEmail, []byte(message))
}

func (m *SMTPPasswordResetMailer) SendTest(ctx context.Context, toEmail string) error {
	if m == nil {
		return nil
	}
	toEmail = strings.TrimSpace(toEmail)
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return fmt.Errorf("parse test recipient: %w", err)
	}

	body := "Hallo,\r\n\r\n" +
		"dies ist eine Test-Mail aus den RailKeeper SMTP-Einstellungen.\r\n\r\n" +
		"Wenn diese Nachricht angekommen ist, kann RailKeeper Passwort-Reset-Links per E-Mail versenden.\r\n"
	message := buildSMTPMessage(m.config.From, toEmail, "RailKeeper SMTP-Test", body)
	return sendSMTP(ctx, m.config, toEmail, []byte(message))
}

func buildPasswordResetMessage(from, toEmail, resetURL, expiresAt string) string {
	body := "Hallo,\r\n\r\n" +
		"für dein RailKeeper-Konto wurde eine Passwort-Rücksetzung angefordert.\r\n\r\n" +
		"Reset-Link:\r\n" + resetURL + "\r\n\r\n" +
		"Der Link ist zeitlich begrenzt und nur einmal nutzbar."
	if expiresAt != "" {
		body += "\r\nGültig bis: " + expiresAt
	}
	body += "\r\n\r\nFalls du die Rücksetzung nicht angefordert hast, kannst du diese E-Mail ignorieren.\r\n"
	return buildSMTPMessage(from, toEmail, "RailKeeper Passwort zurücksetzen", body)
}

func buildSMTPMessage(from, toEmail, subjectText, body string) string {
	subject := mime.QEncoding.Encode("utf-8", subjectText)
	headers := []string{
		"From: " + from,
		"To: " + toEmail,
		"Subject: " + subject,
		"Date: " + time.Now().UTC().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}

func sendSMTP(ctx context.Context, config SMTPPasswordResetMailConfig, toEmail string, message []byte) error {
	addr := net.JoinHostPort(config.Host, config.Port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer func() { _ = conn.Close() }()

	tlsConfig := &tls.Config{ServerName: config.Host, MinVersion: tls.VersionTLS12}
	if config.TLSMode == "implicit" {
		conn = tls.Client(conn, tlsConfig)
	}

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if config.TLSMode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("start smtp tls: %w", err)
			}
		} else {
			return fmt.Errorf("smtp server does not advertise STARTTLS")
		}
	}

	if config.Username != "" || config.Password != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(config.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("smtp recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write smtp message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close smtp message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit smtp: %w", err)
	}
	return nil
}
