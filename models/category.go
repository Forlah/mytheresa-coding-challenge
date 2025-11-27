package models

import "github.com/google/uuid"

type Category struct {
	ID        uint      `gorm:"primaryKey"`
	ProductID uint      `gorm:"not null"`
	Code      uuid.UUID `gorm:"type:uuid;not null"`
	Name      string    `gorm:"not null"`
}

func (c *Category) TableName() string {
	return "product_categories"
}
