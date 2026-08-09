package credential

type credentialModel struct {
	ID           string `gorm:"column:id"`
	ProspectID   string `gorm:"column:prospect_id"`
	Username     string `gorm:"column:username"`
	PasswordHash string `gorm:"column:password_hash"`
}

func (credentialModel) TableName() string {
	return "credentials"
}
