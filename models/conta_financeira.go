package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID        uint `gorm:"primaryKey" param:"account"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" `
	TeamID    uint           `gorm:"index" `
	Name      string
	UnidadeID *uint `gorm:"null"`

	Unidade Unidade `gorm:"foreignKey:UnidadeID"`
}
