package models

import (
	"time"

	"gorm.io/gorm"
)

type ContaFinanceira struct {
	ID        uint `gorm:"primaryKey" param:"unidade"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	TeamID    uint           `gorm:"index"`
	Nome      string         `json:"nome"`
	UnidadeID *uint          `gorm:"null" json:"unidade_id"`

	Unidade Unidade `gorm:"foreignKey:UnidadeID" json:"-"`
}
