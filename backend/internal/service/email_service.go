package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"

	"my-blog/backend/internal/config"
)

var ErrSMTPNotConfigured = errors.New("smtp not configured")

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) IsConfigured() bool {
	return s.cfg.SMTPHost != "" &&
		s.cfg.SMTPUser != "" &&
		s.cfg.SMTPPass != "" &&
		s.cfg.SMTPFrom != ""
}

func (s *EmailService) SendVerifyCode(toEmail, code string) error {
	if !s.IsConfigured() {
		return ErrSMTPNotConfigured
	}

	subject := "【风雨世界】邮箱验证码"
	body := fmt.Sprintf("您的注册验证码为：%s，10分钟内有效，请勿泄露。", code)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.cfg.SMTPFrom, toEmail, subject, body,
	)

	// 465 端口使用 SSL/TLS 直连；587 端口使用 STARTTLS
	if s.cfg.SMTPPort == "465" {
		return s.sendWithTLS(toEmail, msg)
	}
	return s.sendWithStartTLS(toEmail, msg)
}

// sendWithTLS 用于 465 端口（QQ、163 等国内邮箱常用）
func (s *EmailService) sendWithTLS(toEmail, msg string) error {
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	tlsCfg := &tls.Config{
		ServerName: s.cfg.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("连接邮件服务器失败(%s): %w", addr, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp 认证失败(请检查 SMTP_USER/SMTP_PASS 授权码): %w", err)
	}
	if err = client.Mail(s.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("smtp rcpt to %s: %w", toEmail, err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = fmt.Fprint(wc, msg); err != nil {
		_ = wc.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = wc.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return client.Quit()
}

// sendWithStartTLS 用于 587 端口（Gmail、QQ 备用端口等）
func (s *EmailService) sendWithStartTLS(toEmail, msg string) error {
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = s.cfg.SMTPHost
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("连接邮件服务器失败(%s): %w", addr, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: host}
		if err = client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp 认证失败(请检查 SMTP_USER/SMTP_PASS 授权码): %w", err)
	}
	if err = client.Mail(s.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("smtp rcpt to %s: %w", toEmail, err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = fmt.Fprint(wc, msg); err != nil {
		_ = wc.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = wc.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return client.Quit()
}
