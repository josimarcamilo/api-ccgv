package models

import (
	"time"

	"gorm.io/gorm"
)

// Roles admin, cc-secretaria, cc-tesouraria, cc-presidente, cc-fiscal
// é melhor criar o sistema por permissoes e deixar o usuário criar os perfis e selecionar as permissoes
type User struct {
	gorm.Model
	Name     string `gorm:"size:255;not null" json:"name" form:"name"`
	Email    string `gorm:"size:255;unique;not null" json:"email" form:"email"`
	Password string `gorm:"size:255;not null" json:"password" form:"password"`
	TeamID   uint   `gorm:"index"` // FK para Time
	Team     Team   `gorm:"foreignKey:TeamID"`
	Role     string `gorm:"size:50;null" json:"role" form:"role"`
}

type Team struct {
	ID         uint       `gorm:"primaryKey"`
	Name       string     `gorm:"size:255;not null;unique"`
	UserID     uint       `gorm:"index"`
	Users      []User     `gorm:"foreignKey:TeamID"`
	Categories []Category `gorm:"foreignKey:TeamID"`
	Accounts   []Account  `gorm:"foreignKey:TeamID"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Category struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:255;not null" form:"name" json:"name"`
	TeamID    uint           `gorm:"index" json:"team_id"` // FK para Time
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Account struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name" form:"name"`
	Balance   float32        `gorm:"not null;default:0" json:"balance,string"`
	TeamID    uint           `gorm:"index"` // FK para Time
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Transaction struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	TeamID      uint           `gorm:"index;index:idx_date_team"`
	Date        string         `gorm:"type:date;index:idx_date_team" json:"date_at" form:"date_at"`
	Type        int            `gorm:"not null" json:"type" form:"type"` // 1 - Entrada, 2 - Saída, 3 - Transferência
	Description string         `gorm:"size:255;not null" json:"description" form:"description"`
	Value       float64        `gorm:"not null" json:"value" form:"value"`
	CategoryID  uint           `gorm:"not null" json:"category_id" form:"category_id"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category"`
	AccountID   uint           `gorm:"not null" json:"account_id" form:"account_id"`
	Account     Account        `gorm:"foreignKey:AccountID" json:"account"`
	Proof       *string        `json:"proof" form:"proof"` // Caminho do arquivo
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
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
	Observation   string
}
