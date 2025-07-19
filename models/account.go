package models

import (
	"time"

	"gorm.io/gorm"
)

type AccountFilter struct {
	Virtual string `query:"virtual"`
}

type Account struct {
	ID                  uint `gorm:"primaryKey" param:"account"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
	TeamID              uint           `gorm:"index"`
	Name                string
	Virtual             bool          `gorm:"not null; default:false"`
	RealTransactions    []Transaction `gorm:"foreignKey:AccountID;references:ID"`
	VirtualTransactions []Transaction `gorm:"foreignKey:AccountVirtualID;references:ID"`
}
