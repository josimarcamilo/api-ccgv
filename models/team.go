package models

import (
	"time"

	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	Name       string     `gorm:"size:255;not null;"`
	UserID     uint       `gorm:"index"`
	Categories []Category `gorm:"foreignKey:TeamID"`
	Accounts   []Account  `gorm:"foreignKey:TeamID"`
}

type Role struct {
	ID        uint   `gorm:"primaryKey"`
	TeamID    uint   `gorm:"index"`
	Name      string `gorm:"size:255;not null;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserTeam struct {
	ID        uint `gorm:"primaryKey"`
	TeamID    uint `gorm:"index"`
	UserID    uint `gorm:"index"`
	Team      Team `gorm:"foreignKey:TeamID"`
	User      User `gorm:"foreignKey:UserID"`
	RoleId    uint `gorm:""`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
