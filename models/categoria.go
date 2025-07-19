package models

import (
	"time"

	"gorm.io/gorm"
)

const CATEGORY_ENTRY = 1
const CATEGORY_EXIT = 2

type Category struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	TeamID          uint           `gorm:"index" json:"team_id"`
	Type            int8           `gorm:"null" json:"type"`
	Name            string         `gorm:"size:255;not null" form:"name" json:"name"`
	UseMap          bool           `gorm:"null;default:false" form:"use_map" json:"use_map"`
	TipoRepasse     int8           `gorm:"null;default:0" form:"tipo_repasse" json:"tipo_repasse"` // 0 - não repassa, 1 - 10%, 2 - 2,5%
	TransactionsMap []Transaction  `gorm:"foreignKey:CategoryMapID;references:ID" json:"transactions"`
}
