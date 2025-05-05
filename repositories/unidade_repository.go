package repositories

import (
	"jc-financas/models"
)

func CreateUnidade(model *models.Unidade) error {
	return DB.Create(model).Error
}

func GetUnidade(unidadeId, teamID uint) (*models.Unidade, error) {
	var model models.Unidade
	if err := DB.Where("team_id = ?", teamID).Where("id = ?", unidadeId).First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func UpdateUnidade(model *models.Unidade) error {
	return DB.Save(&model).Error
}

func GetUnidades(teamID uint) ([]models.Unidade, error) {
	var models []models.Unidade

	if err := DB.Where("team_id = ?", teamID).Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
