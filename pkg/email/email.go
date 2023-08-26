package email

import (
	"fmt"
	"net/smtp"

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

func (s *EmailService) SendEmail(config domain.SendEmailConfig, toEmail string) error {
	e := email.NewEmail()

	e.From = fmt.Sprintf("%s <%s>", s.Name, s.Email)
	e.Subject = config.Subject
	e.HTML = []byte(config.Content)
	e.To = []string{toEmail}

	smtpAuth := smtp.PlainAuth("", s.Email, s.Password, s.Host)
	return e.Send(fmt.Sprintf("%s:%d", s.Host, s.Port), smtpAuth)
}
