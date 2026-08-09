package prospect

type prospectModel struct {
	ID       string `gorm:"column:id"`
	Username string `gorm:"column:username"`
	Email    string `gorm:"column:email"`
	Status   string `gorm:"column:status"`
}

func (prospectModel) TableName() string {
	return "prospects"
}
