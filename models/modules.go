package models

type Module struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"size:255;not null" json:"name" form:"name"`
	Description string `gorm:"size:255;not null" json:"description" form:"description"`
	Icon        string `gorm:"size:255;not null" json:"icon" form:"icon"`
}

type ModulePermission struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	ModuleID   uint   `gorm:"index" json:"module_id" form:"module_id"`
	Permission string `gorm:"size:255;not null" json:"permission" form:"permission"`
}
