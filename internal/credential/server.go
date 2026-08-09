package credential

import (
	"context"

	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CredentialServer struct {
	credentialv1.UnimplementedCredentialServiceServer

	Repo   Repository
	Hasher Hasher
}

func (s *CredentialServer) CreateCredential(ctx context.Context, req *credentialv1.CreateCredentialRequest) (*credentialv1.CreateCredentialResponse, error) {
	passwordHash, err := s.Hasher.Hash(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not hash password")
	}

	credentialId, err := s.Repo.Insert(ctx, req.ProspectId, req.Username, passwordHash)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not create credential")
	}
	return &credentialv1.CreateCredentialResponse{CredentialId: credentialId}, nil
}
