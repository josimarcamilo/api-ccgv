package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:255;not null" form:"name" json:"name"`
	TeamID    uint           `gorm:"index" json:"team_id"` // FK para Time
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// categorias para estarem igual ao mapa mensal
type CategoryMap struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:255;not null" form:"name" json:"name"`
	TeamID    uint           `gorm:"index" json:"team_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Account struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name" form:"name"`
	Balance     float64        `gorm:"type:decimal(10,2);not null;default:0" json:"balance,string"`
	BalanceDate string         `json:"balance_date"`
	TeamID      uint           `gorm:"index"` // FK para Time
	ToReceive   bool           `gorm:"default:false" form:"to_receive" json:"to_receive"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Transaction struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	TeamID            uint           `gorm:"index:idx_date_team;index:idx_external_id"`
	AccountID         uint           `gorm:"null" json:"account_id" form:"account_id"`
	Date              string         `gorm:"type:date;index:idx_date_team" json:"date_at" form:"date"`
	Type              int            `gorm:"not null" json:"type" form:"type"` // 1 - Entrada, 2 - Saída, 3 - Transferência
	Description       string         `gorm:"size:255;not null" json:"description" form:"description"`
	Value             float64        `gorm:"not null" json:"value" form:"value"`
	CategoryID        uint           `gorm:"null" json:"category_id" form:"category_id"`
	Category          Category       `gorm:"foreignKey:CategoryID" json:"category"`
	CategoryMapID     uint           `gorm:"null" json:"category_map_id" form:"category_map_id"`
	CategoryMap       CategoryMap    `gorm:"foreignKey:CategoryMapID" json:"category_map"`
	Account           Account        `gorm:"foreignKey:AccountID" json:"account"`
	Proof             *string        `json:"proof" form:"proof"`
	TransactionOrigin *uint          `gorm:"null" json:"transaction_origin"`
	Transfer          bool           `gorm:"not null;default:false" form:"transfer" json:"transfer"`
	ExternalId        string         `gorm:"null;index:idx_external_id" json:"external_id" form:"external_id"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// comprovantes
type Proof struct {
	gorm.Model
	TransactionID uint   `gorm:"index"`             // FK para Transação
	FilePath      string `gorm:"size:500;not null"` // Caminho do arquivo
	UploadedAt    time.Time
}

// aprovacoes tesoureiros e conselho fiscal
type Approval struct {
	gorm.Model
	TransactionID uint   `gorm:"index"` // FK para Transação
	UserID        uint   `gorm:"index"` // FK para usuário
	Status        string // aprovado, reprovado
	Observation   *string
}
