package main

import (
	"log"
	"net/http"

	"github.com/minhloi0901/go-password-regis/internal/config"
	"github.com/minhloi0901/go-password-regis/internal/email"
	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	"github.com/minhloi0901/go-password-regis/internal/prospect"
	"github.com/minhloi0901/go-password-regis/internal/registration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// start postgre connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// start credential grpc connection
	credentialConn, err := grpc.NewClient(
		cfg.CredentialServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // tls connection
	)
	if err != nil {
		log.Fatalf("cannot connect grpc port: %s with error %v", cfg.CredentialServiceAddr, err)
	}
	defer credentialConn.Close()

	// start email service connection
	emailService := email.NewGomailEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	rh := registration.NewRegisterHandler(
		prospect.NewPostgresRespository(db),
		credentialv1.NewCredentialServiceClient(credentialConn),
		emailService,
	)

	// start registration http
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: registration.NewRouter(rh),
	}
	log.Println("Registration Server is running on port: ", cfg.HTTPAddr)
	log.Fatal(server.ListenAndServe())
}
