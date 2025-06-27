package repositories

import (
	"jc-financas/models"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func CreateCategory(model *models.Category) error {
	return DB.Create(model).Error
}

func UpsertCategory(model *models.Category) (*models.Category, error) {
	var existing models.Category

	// Tenta encontrar a categoria pelo nome (ou outro campo único)
	err := DB.Where("name = ?", model.Name).
		Where("type = ?", model.Type).
		Where("team_id = ?", model.TeamID).
		Where("use_map = ?", model.UseMap).
		First(&existing).Error

	if err == nil {
		return &existing, nil
		// Já existe: atualiza os campos necessários, se quiser
		// existing.Description = model.Description // exemplo
		// return DB.Save(&existing).Error
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Não existe: cria nova
		create := DB.Create(model)
		return model, create.Error
	}

	// Outro erro (problema de conexão, etc)
	return nil, err
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
