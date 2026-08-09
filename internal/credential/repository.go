package credential

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Insert(ctx context.Context, prospectID, username, passwordHash string) (credentialID string, err error)
}

type PostgresRepository struct {
	// Pool *pgxpool.Pool
	DB *gorm.DB
}

func NewPostgresRespository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{DB: db}
}

func (r *PostgresRepository) Insert(ctx context.Context, prospectID, username, passwordHash string) (string, error) {
	credentialID := NewCredentialID()

	model := credentialModel{
		ID:           credentialID,
		ProspectID:   prospectID,
		Username:     username,
		PasswordHash: passwordHash,
	}

	if err := r.DB.WithContext(ctx).Create(&model).Error; err != nil {
		return "", err
	}

	return credentialID, nil
}

func NewCredentialID() string {
	return uuid.NewString()
}
