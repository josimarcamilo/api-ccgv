package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	TransactionTypeEntrada = 1
	TransactionTypeSaida   = 2
)

type Transaction struct {
	ID        uint `gorm:"primarykey" param:"transaction"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	TeamID           uint  `gorm:"index:idx_date_team;index:idx_external_id"`
	AccountID        *uint `json:"AccountID"`
	AccountVirtualID *uint `json:"AccountVirtualID"`
	CategoryID       *uint `json:"CategoryID"`
	CategoryMapID    *uint `json:"CategoryMapID"`

	Type        int     `gorm:"not null"` // 1 - Entrada, 2 - Saída
	IsTransfer  bool    `gorm:"not null; default:false"`
	Date        string  `gorm:"type:date;index:idx_date_team"`
	Description string  `gorm:"size:255;not null"`
	Value       float64 `gorm:"not null"`

	TransactionOriginId *uint
	ExternalId          *string `gorm:"index:idx_external_id"`
	ReceiptUrl          *string

	Account        Account  `gorm:"foreignKey:AccountID"`
	AccountVirtual Account  `gorm:"foreignKey:AccountVirtualID"`
	Category       Category `gorm:"foreignKey:CategoryID"`
	CategoryMap    Category `gorm:"foreignKey:CategoryMapID"`
}
