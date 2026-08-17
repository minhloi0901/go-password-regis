package prospect

import "time"

type prospectModel struct {
	ID               string    `gorm:"column:id"`
	Username         string    `gorm:"column:username"`
	Email            string    `gorm:"column:email"`
	Status           string    `gorm:"column:status"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	ExpiresAt        time.Time `gorm:"column:expires_at"`
	VerificationCode string    `gorm:"column:verification_code"`
	CodeExpiresAt    time.Time `gorm:"column:code_expires_at"`
}

func (prospectModel) TableName() string {
	return "prospects"
}
