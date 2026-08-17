package email

import (
	"context"
	"log"
	"time"

	gomail "gopkg.in/gomail.v2"
)

type EmailService interface {
	SendVerificationEmail(ctx context.Context, email, code string) error
}

type GomailEmailService struct {
	from    string
	dialer  *gomail.Dialer
	codeTTL time.Duration
}

func NewGomailEmailService(host string, port int, username, password, from string, codeTTL time.Duration) *GomailEmailService {
	return &GomailEmailService{
		from:    from,
		dialer:  gomail.NewDialer(host, port, username, password),
		codeTTL: codeTTL,
	}
}

// not implement
func (s *GomailEmailService) SendVerificationEmail(ctx context.Context, email, code string) error {
	log.Printf("Send verificaton code to %s: %s", email, code)

	return nil
}
