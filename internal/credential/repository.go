package credential

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Insert(ctx context.Context, prospectID, username, passwordHash string) (string, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	FindByUsername(ctx context.Context, username string) (Credential, error)
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
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return "", ErrUsernameTaken
		}
		return "", err
	}

	return credentialID, nil
}

func (r *PostgresRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var model credentialModel

	if err := r.DB.WithContext(ctx).Where("username = ?", username).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *PostgresRepository) FindByUsername(ctx context.Context, username string) (Credential, error) {
	var model credentialModel

	if err := r.DB.WithContext(ctx).Where("username = ?", username).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Credential{}, ErrCredentialNotFound
		}
		return Credential{}, err
	}

	return Credential{
		ID:           model.ID,
		ProspectID:   model.ProspectID,
		Username:     model.Username,
		PasswordHash: model.PasswordHash,
	}, nil
}

func NewCredentialID() string {
	return uuid.NewString()
}
