package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/b0shka/backend/internal/domain"
	"github.com/jordan-wright/email"
)

type EmailService struct {
	Name     string
	Email    string
	Password string
	Host     string
	Port     int
}

func NewEmailService(name, email, password, host string, port int) *EmailService {
	return &EmailService{
		Name:     name,
		Email:    email,
		Password: password,
		Host:     host,
		Port:     port,
	}
}

func (s *EmailService) SendEmailMessage(toEmail, templateFile, subject string, contentData any) error {
	var content bytes.Buffer
	contentHtml, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	err = contentHtml.Execute(&content, contentData)
	if err != nil {
		return err
	}

	config := domain.SendEmailConfig{
		Subject: subject,
		Content: content.String(),
	}

	e := email.NewEmail()

	e.From = fmt.Sprintf("%s <%s>", s.Name, s.Email)
	e.Subject = config.Subject
	e.HTML = []byte(config.Content)
	e.To = []string{toEmail}

	smtpAuth := smtp.PlainAuth("", s.Email, s.Password, s.Host)
	return e.Send(fmt.Sprintf("%s:%d", s.Host, s.Port), smtpAuth)
}
