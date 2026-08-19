package email

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/minhloi0901/go-password-regis/internal/prospect"
	gomail "gopkg.in/gomail.v2"
)

//go:generate mockgen -source=internal/email/service.go -destination=internal/email/mocks/mock_service.go -package=mocks
type EmailService interface {
	SendVerificationEmail(ctx context.Context, email, code string) error
}

type GomailEmailService struct {
	from   string
	dialer *gomail.Dialer
}

func NewGomailEmailService(host string, port int, username, password, from string) *GomailEmailService {
	return &GomailEmailService{
		from:   from,
		dialer: gomail.NewDialer(host, port, username, password),
	}
}

func (s *GomailEmailService) SendVerificationEmail(ctx context.Context, email, code string) error {
	// avoid CRLF injection
	if strings.ContainsAny(email, "\n\r") {
		log.Printf("Register - Send Verfication Email: Contain invalid email")
		return ErrInvalidEmail
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.from)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Verification Code")

	message.SetBody("text/plain", s.buildBody(code))

	if err := s.dialer.DialAndSend(message); err != nil {
		log.Println("Error sending email:", err)
		return err
	}

	log.Println("Email sent to:", email)
	return nil
}

func (s *GomailEmailService) buildBody(code string) string {
	return fmt.Sprintf(
		"Please verify your email address using the code below to complete account setup\n\n"+
			"%s\n\n"+
			"The code expires in %d minutes",
		code,
		int(prospect.VerificationCodeTTL.Minutes()))
}
