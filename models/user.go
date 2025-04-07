package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"size:255;not null" json:"name" form:"name"`
	Email    string `gorm:"size:255;unique;not null" json:"email" form:"email"`
	Password string `gorm:"size:255;not null" json:"password" form:"password"`
	TeamID   uint   `gorm:"index"` // FK para Time
	Team     Team   `gorm:"foreignKey:TeamID"`
	Role     string `gorm:"size:50;null" json:"role" form:"role"`
}
