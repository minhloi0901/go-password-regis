package credential

import (
	"context"
	"errors"

	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CredentialServer struct {
	credentialv1.UnimplementedCredentialServiceServer

	CredentialRepo Repository
	Hasher         Hasher
}

func (s *CredentialServer) CreateCredential(ctx context.Context, req *credentialv1.CreateCredentialRequest) (*credentialv1.CreateCredentialResponse, error) {
	if err := s.validateUniqueCredential(ctx, req.Username); err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			return nil, status.Errorf(codes.AlreadyExists, "username already exists")
		}
		return nil, status.Error(codes.Internal, "could not validate unique credential")
	}

	passwordHash, err := s.Hasher.Hash(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not hash password")
	}

	credentialId, err := s.CredentialRepo.Insert(ctx, req.ProspectId, req.Username, passwordHash)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not create credential")
	}
	return &credentialv1.CreateCredentialResponse{CredentialId: credentialId}, nil
}

func (s *CredentialServer) validateUniqueCredential(ctx context.Context, username string) error {
	exists, err := s.CredentialRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return err
	}
	if exists {
		return ErrUsernameTaken
	}

	return nil
}
