package prospect

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Insert(ctx context.Context, username, email string) (string, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	DeleteById(ctx context.Context, id string) error
	FindById(ctx context.Context, id string) (Prospect, error)
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
		Status:   StatusPending,
	}

	if err := r.DB.WithContext(ctx).Create(&model).Error; err != nil {
		// log.Printf("insert prospect failed: %v", err)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return "", ErrProspectConflict
		}
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

func (r *PostgresRepository) DeleteById(ctx context.Context, id string) error {
	var model prospectModel
	if err := r.DB.WithContext(ctx).Where("id = ?", id).Delete(&model).Error; err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) FindById(ctx context.Context, id string) (Prospect, error) {
	var model prospectModel

	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Prospect{}, ErrProspectNotFound
		}
		return Prospect{}, err
	}

	return Prospect{
		ID:       model.ID,
		Username: model.Username,
		Email:    model.Email,
		Status:   model.Status,
	}, nil
}

func NewProspectID() string {
	return uuid.NewString()
}
