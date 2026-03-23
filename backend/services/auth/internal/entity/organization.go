package entity

type Organization struct {
	ID string `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex;not null"`
	URL string `gorm:"uniqueIndex;not null"`
}