package credential

import (
	"context"
	"errors"
	"log"

	credentialv1 "github.com/minhloi0901/go-password-regis/internal/genproto/credential/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CredentialServer struct {
	credentialv1.UnimplementedCredentialServiceServer

	CredentialRepo Repository
	Hasher         Hasher
}

func testDeleteProspect(username string) bool {
	return username == "force-fail"
}

func (s *CredentialServer) CreateCredential(ctx context.Context, req *credentialv1.CreateCredentialRequest) (*credentialv1.CreateCredentialResponse, error) {
	if testDeleteProspect(req.Username) {
		return nil, status.Error(codes.Internal, "forced create credential fail")
	}

	if err := s.validateUniqueCredential(ctx, req.Username); err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		}
		log.Printf("CreateCredential - Check Unique: %v", err)
		return nil, status.Error(codes.Internal, "could not validate unique credential")
	}

	passwordHash, err := s.Hasher.Hash(req.Password)
	if err != nil {
		log.Printf("CreateCredential - Hash Password: %v", err)
		return nil, status.Error(codes.Internal, "could not hash password")
	}

	credentialId, err := s.CredentialRepo.Insert(ctx, req.ProspectId, req.Username, passwordHash)
	if err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		}
		log.Printf("CreateCredential - Insert Credential: %v", err)
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

// not implement token yet
func (s *CredentialServer) VerifyCredential(ctx context.Context, req *credentialv1.VerifyCredentialRequest) (*credentialv1.VerifyCredentialResponse, error) {
	cred, err := s.CredentialRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			return &credentialv1.VerifyCredentialResponse{
				Valid: false,
			}, nil
		}
		return nil, status.Error(codes.Internal, "could not find credential by username")
	}

	match, err := s.Hasher.Compare(cred.PasswordHash, req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not compare hash password")
	}
	if !match {
		return &credentialv1.VerifyCredentialResponse{
			Valid: false,
		}, nil
	}

	return &credentialv1.VerifyCredentialResponse{
		Valid:      true,
		ProspectId: cred.ProspectID,
	}, nil
}
