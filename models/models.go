package models

import (
	"time"

	"gorm.io/gorm"
)

// Roles admin, cc-secretaria, cc-tesouraria, cc-presidente, cc-fiscal
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
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255;not null"`
	TeamID    uint   `gorm:"index"` // FK para Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Account struct {
	gorm.Model
	Name    string  `gorm:"size:255;not null"`
	Balance float64 `gorm:"not null;default:0.0"`
	TeamID  uint    `gorm:"index"` // FK para Time
}

type Transaction struct {
	gorm.Model
	Type          string  `gorm:"size:50;not null"` // Ex: "entrada 1", "saída 2", "transferência 3"
	Amount        float64 `gorm:"not null"`
	Description   string  `gorm:"size:500"`
	BankAccountID uint    `gorm:"index"`                    // FK para Conta Bancária
	CategoryID    uint    `gorm:"index"`                    // FK para Categoria
	TeamID        uint    `gorm:"index"`                    // FK para Time
	Proofs        []Proof `gorm:"foreignKey:TransactionID"` // Relacionamento com comprovantes
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
