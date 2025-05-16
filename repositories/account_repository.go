package repositories

import (
	"jc-financas/models"
)

func CreateCategory(model *models.Category) error {
	return DB.Create(model).Error
}

func GetCategory(id, teamID uint) (*models.Category, error) {
	var model models.Category
	if err := DB.Where("team_id = ?", teamID).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func UpdateCategory(model *models.Category) error {
	return DB.Save(&model).Error
}

func GetCategorys(teamID uint) ([]models.Category, error) {
	var models []models.Category

	if err := DB.Where("team_id = ?", teamID).Preload("Unidade").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
