package services

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	AppURL   string
}

func SendMagicLinkEmail(cfg EmailConfig, toEmail, token string) error {
	link := fmt.Sprintf("%s/verify?token=%s", cfg.AppURL, token)

	htmlBody, err := buildMagicLinkHTML(link)
	if err != nil {
		return fmt.Errorf("failed to build email html: %w", err)
	}

	subject := "Secure link to log in to Seismic.icu"
	message := buildMIMEMessage(cfg.Username, toEmail, subject, htmlBody)

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return smtp.SendMail(addr, auth, cfg.Username, []string{toEmail}, []byte(message))
}

func buildMagicLinkHTML(link string) (string, error) {
	content, err := os.ReadFile("services/templates/magic_link.html")
	if err != nil {
		return "", err
	}

	html := strings.ReplaceAll(string(content), "{{MAGIC_LINK_URL}}", link)
	return html, nil
}

func buildMIMEMessage(from, to, subject, htmlBody string) string {
	return fmt.Sprintf(
		"From: Seismic <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"%s",
		from, to, subject, htmlBody,
	)
}
