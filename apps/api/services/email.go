package services

import (
	"fmt"
	"html"
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
	baseURL := appURL(cfg)
	link := fmt.Sprintf("%s/verify?token=%s", baseURL, token)

	htmlBody, err := buildActionEmailHTML(
		cfg,
		"Let's get you signed in",
		"Use the secure link below to continue to your Seismic dashboard.",
		"Sign in to Seismic",
		link,
	)
	if err != nil {
		return fmt.Errorf("failed to build email html: %w", err)
	}

	subject := "Secure link to log in to Seismic.icu"
	message := buildMIMEMessage(cfg.Username, toEmail, subject, htmlBody)

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return smtp.SendMail(addr, auth, cfg.Username, []string{toEmail}, []byte(message))
}

func buildActionEmailHTML(cfg EmailConfig, title, body, ctaLabel, actionURL string) (string, error) {
	content, err := os.ReadFile("services/templates/magic_link.html")
	if err != nil {
		return "", err
	}

	baseURL := appURL(cfg)
	rendered := string(content)
	rendered = strings.ReplaceAll(rendered, "{{APP_URL}}", html.EscapeString(baseURL))
	rendered = strings.ReplaceAll(rendered, "{{LOGO_URL}}", html.EscapeString(logoURL(baseURL)))
	rendered = strings.ReplaceAll(rendered, "{{EMAIL_TITLE}}", html.EscapeString(title))
	rendered = strings.ReplaceAll(rendered, "{{EMAIL_BODY}}", html.EscapeString(body))
	rendered = strings.ReplaceAll(rendered, "{{CTA_LABEL}}", html.EscapeString(ctaLabel))
	rendered = strings.ReplaceAll(rendered, "{{ACTION_URL}}", html.EscapeString(actionURL))
	return rendered, nil
}

func appURL(cfg EmailConfig) string {
	return strings.TrimRight(cfg.AppURL, "/")
}

func logoURL(baseURL string) string {
	return baseURL + "/images/seismic-logo.png"
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

func SendGoalReminderEmail(cfg EmailConfig, toEmail, goalLabel, progress, target string, progressPercent int, period string) error {
	content, err := os.ReadFile("services/templates/goal_reminder.html")
	if err != nil {
		return err
	}

	rendered := string(content)
	baseURL := appURL(cfg)
	rendered = strings.ReplaceAll(rendered, "{{APP_URL}}", html.EscapeString(baseURL))
	rendered = strings.ReplaceAll(rendered, "{{LOGO_URL}}", html.EscapeString(logoURL(baseURL)))
	rendered = strings.ReplaceAll(rendered, "{{GOAL_LABEL}}", html.EscapeString(goalLabel))
	rendered = strings.ReplaceAll(rendered, "{{PROGRESS}}", html.EscapeString(progress))
	rendered = strings.ReplaceAll(rendered, "{{TARGET}}", html.EscapeString(target))
	rendered = strings.ReplaceAll(rendered, "{{PROGRESS_PERCENT}}", html.EscapeString(fmt.Sprintf("%d", progressPercent)))
	rendered = strings.ReplaceAll(rendered, "{{PERIOD}}", html.EscapeString(period))

	subject := "You're falling behind on your Seismic goal"
	message := buildMIMEMessage(cfg.Username, toEmail, subject, rendered)

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return smtp.SendMail(addr, auth, cfg.Username, []string{toEmail}, []byte(message))
}

func SendEmailChangeConfirmation(cfg EmailConfig, newEmail, token string) error {
	baseURL := appURL(cfg)
	link := fmt.Sprintf("%s/confirm-email?token=%s", baseURL, token)

	htmlBody, err := buildActionEmailHTML(
		cfg,
		"Confirm your new email",
		"Confirm this address so Seismic can use it for future sign-ins.",
		"Confirm email",
		link,
	)
	if err != nil {
		return fmt.Errorf("failed to build email html: %w", err)
	}

	subject := "Confirm your new email for Seismic"
	message := buildMIMEMessage(cfg.Username, newEmail, subject, htmlBody)

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return smtp.SendMail(addr, auth, cfg.Username, []string{newEmail}, []byte(message))
}
