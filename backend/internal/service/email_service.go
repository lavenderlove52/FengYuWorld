package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"

	"my-blog/backend/internal/config"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) SendVerifyCode(toEmail, code string) error {
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
		InsecureSkipVerify: false,
		ServerName:         s.cfg.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err = client.Mail(s.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	defer wc.Close()
	_, err = fmt.Fprint(wc, msg)
	return err
}

// sendWithStartTLS 用于 587 端口（Gmail 等）
func (s *EmailService) sendWithStartTLS(toEmail, msg string) error {
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	host, _, _ := net.SplitHostPort(addr)
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, host)
	return smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{toEmail}, []byte(msg))
}
