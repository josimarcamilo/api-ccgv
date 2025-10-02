package repositories

import (
	"jc-financas/models"
)

func CreateTransaction(model *models.Transaction) error {
	return DB.Create(model).Error
}

func GetTransaction(ID, teamID uint) (*models.Transaction, error) {
	var model models.Transaction
	if err := DB.Where("team_id = ?", teamID).Where("id = ?", ID).Preload("Account").Preload("AccountVirtual").Preload("Category").Preload("CategoryMap").First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func UpdateTransaction(model *models.Transaction) error {
	return DB.Model(&models.Transaction{}).
		Where("id = ?", model.ID).
		Select("AccountID", "AccountVirtualID", "CategoryID", "CategoryMapID", "Date", "Description", "Type", "Value", "IsTransfer").
		Updates(model).Error
}

func DeleteTransaction(model *models.Transaction) error {
	return DB.Model(&models.Transaction{}).
		Where("id = ?", model.ID).
		Delete(model).Error
}

// func BuildTransactionUpdateMap(tx *models.Transaction) map[string]interface{} {
// 	updates := make(map[string]interface{})

// 	if tx.AccountID.Valid {
// 		updates["account_id"] = tx.AccountID.Int64
// 	} else {
// 		updates["account_id"] = nil
// 	}

// 	if tx.UnidadeID.Valid {
// 		updates["unidade_id"] = tx.UnidadeID.Int64
// 	} else {
// 		updates["unidade_id"] = nil
// 	}

// 	if tx.CategoryID.Valid {
// 		updates["category_id"] = tx.CategoryID.Int64
// 	} else {
// 		updates["category_id"] = nil
// 	}

// 	if tx.CategoryMapID.Valid {
// 		updates["category_map_id"] = tx.CategoryMapID.Int64
// 	} else {
// 		updates["category_map_id"] = nil
// 	}

// 	if tx.TransactionOriginId.Valid {
// 		updates["transaction_origin_id"] = tx.TransactionOriginId.Int64
// 	} else {
// 		updates["transaction_origin_id"] = nil
// 	}

// 	if tx.ExternalId.Valid {
// 		updates["external_id"] = tx.ExternalId.String
// 	} else {
// 		updates["external_id"] = nil
// 	}

// 	if tx.ReceiptUrl.Valid {
// 		updates["receipt_url"] = tx.ReceiptUrl.String
// 	} else {
// 		updates["receipt_url"] = nil
// 	}

// 	// Campos obrigatórios (sem nullables)
// 	updates["type"] = tx.Type
// 	updates["is_transfer"] = tx.IsTransfer
// 	updates["date"] = tx.Date
// 	updates["description"] = tx.Description
// 	updates["value"] = tx.Value

// 	return updates
// }

func GetTransactions(teamID uint, filter models.TransactionFilter) ([]models.Transaction, error) {
	var models []models.Transaction
	query := DB.Where("team_id = ?", teamID)

	if filter.Type != "" {
		query.Where("type = ?", filter.Type)
	}

	if filter.StartDate != "" {
		query.Where("date >= ?", filter.StartDate)
	}

	if filter.EndDate != "" {
		query.Where("date <= ?", filter.EndDate)
	}

	if filter.AccountID != "" {
		query.Where("account_id = ?", filter.AccountID)
	}

	if filter.AccountVirtualID != "" {
		query.Where("account_virtual_id = ?", filter.AccountVirtualID)
	}

	if filter.CategoryID != "" && filter.CategoryID != "sem_categoria" {
		query.Where("category_id = ?", filter.CategoryID)
	}

	if filter.CategoryID == "sem_categoria" {
		query.Where("category_id IS NULL")
	}

	if filter.OrderDate != "" {
		query.Order("date " + filter.OrderDate)
	} else {
		query.Order("date DESC")
	}

	if err := query.Preload("Account").Preload("AccountVirtual").Preload("Category").Preload("CategoryMap").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
