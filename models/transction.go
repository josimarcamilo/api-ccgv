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

	Type        int8   `gorm:"not null"` // 1 - Entrada, 2 - Saída
	IsTransfer  bool   `gorm:"not null; default:false"`
	Date        string `gorm:"type:date;index:idx_date_team"`
	Description string `gorm:"size:255;not null"`
	Value       int    `gorm:"not null"`

	TransactionOriginId *uint
	ExternalId          *string `gorm:"index:idx_external_id"`
	ReceiptUrl          *string

	Account        Account  `gorm:"foreignKey:AccountID"`
	AccountVirtual Account  `gorm:"foreignKey:AccountVirtualID"`
	Category       Category `gorm:"foreignKey:CategoryID"`
	CategoryMap    Category `gorm:"foreignKey:CategoryMapID"`
}

type TransactionList struct {
	ID                 uint
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Type               int8
	IsTransfer         bool
	Date               string
	Description        string
	Value              int
	ReceiptUrl         *string
	AccountID          uint
	AccountName        string
	AccountVirtualID   uint
	AccountVirtualName string
	CategoryID         uint
	CategoryName       string
	CategoryMapID      uint
	CategoryMapName    string
}

type TransactionFilter struct {
	Type             string `query:"type"`
	StartDate        string `query:"start_date"`
	EndDate          string `query:"end_date"`
	AccountID        string `query:"account_id"`
	AccountVirtualID string `query:"account_virtual_id"`
}
