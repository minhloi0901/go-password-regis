package main

import (
	"log"
	"net"

	"github.com/minhloi0901/go-password-regis/internal/config"
	"github.com/minhloi0901/go-password-regis/internal/credential"
	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("crendential service is runnin....")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	credentialServer := &credential.CredentialServer{
		Repo:   credential.NewPostgresRespository(db),
		Hasher: credential.NewBcryptHasher(),
	}

	grpcServer := grpc.NewServer()
	credentialv1.RegisterCredentialServiceServer(grpcServer, credentialServer)

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("Failed to listen gRPC: %v", err)
	}

	log.Println("Credential Service is running on port: ", cfg.GRPCAddr)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to server gRPC: %v", err)
	}
}
