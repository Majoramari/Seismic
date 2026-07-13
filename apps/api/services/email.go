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

func SendGoalReminderEmail(cfg EmailConfig, toEmail, goalLabel, progress, target, period string) error {
	content, err := os.ReadFile("services/templates/goal_reminder.html")
	if err != nil {
		return err
	}

	html := string(content)
	html = strings.ReplaceAll(html, "{{GOAL_LABEL}}", goalLabel)
	html = strings.ReplaceAll(html, "{{PROGRESS}}", progress)
	html = strings.ReplaceAll(html, "{{TARGET}}", target)
	html = strings.ReplaceAll(html, "{{PERIOD}}", period)
	html = strings.ReplaceAll(html, "{{APP_URL}}", cfg.AppURL)

	subject := "You're falling behind on your Seismic goal"
	message := buildMIMEMessage(cfg.Username, toEmail, subject, html)

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return smtp.SendMail(addr, auth, cfg.Username, []string{toEmail}, []byte(message))
}

func SendEmailChangeConfirmation(cfg EmailConfig, newEmail, token string) error {
	link := fmt.Sprintf("%s/confirm-email?token=%s", cfg.AppURL, token)

	htmlBody, err := buildMagicLinkHTML(link) // reuse the same clean template
	if err != nil {
		return fmt.Errorf("failed to build email html: %w", err)
	}

	subject := "Confirm your new email for Seismic"
	message := buildMIMEMessage(cfg.Username, newEmail, subject, htmlBody)

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return smtp.SendMail(addr, auth, cfg.Username, []string{newEmail}, []byte(message))
}
