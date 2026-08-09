package prospect

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken    = errors.New("email already used")
	ErrUsernameTaken = errors.New("username already used")
)

type Repository interface {
	Insert(ctx context.Context, username, email string) (prospectID string, err error)
}

type PostgresRepository struct {
	// Pool *pgxpool.Pool
	DB *gorm.DB
}

func NewPostgresRespository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{DB: db}
}

func (r *PostgresRepository) Insert(ctx context.Context, username, email string) (string, error) {
	prospectId := NewProspectID()

	model := prospectModel{
		ID:       prospectId,
		Username: username,
		Email:    email,
		Status:   "pending",
	}

	if err := r.DB.WithContext(ctx).Create(&model).Error; err != nil {
		log.Printf("insert prospect failed: %v", err)
		return "", err
	}

	return prospectId, nil
}

func (r *PostgresRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var model prospectModel

	if err := r.DB.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *PostgresRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var model prospectModel

	if err := r.DB.WithContext(ctx).Where("username = ?", username).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func NewProspectID() string {
	return uuid.NewString()
}
