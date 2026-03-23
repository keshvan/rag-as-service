package entity

type User struct {
	GUID     string `gorm:"primaryKey"`
	Role string `gorm:"not null"` // роли только owner, admin, member 
	OrganizationID string `gorm:"not null;index"`
	Email    string `gorm:"uniqueIndex;not null"`
	PassHash []byte
}
