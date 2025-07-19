package repositories

import (
	"jc-financas/models"

	"gorm.io/gorm"
)

func CreateAccount(model *models.Account) error {
	return DB.Create(model).Error
}

func GetAccount(id, teamID uint) (*models.Account, error) {
	var model models.Account
	if err := DB.Where("team_id = ?", teamID).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func UpdateAccount(model *models.Account) error {
	return DB.Save(&model).Error
}

func GetAccounts(teamID uint, filter models.AccountFilter) ([]models.Account, error) {
	var models []models.Account
	query := DB.Where("team_id = ?", teamID)
	if filter.Virtual != "" {
		query.Where("virtual = ?", filter.Virtual)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func GetAllBalances(endDate string, teamId uint) AccountBalanceReport {
	var accountBalances []AccountBalance
	DB.Model(&models.Transaction{}).
		Joins("JOIN accounts ON accounts.id = transactions.account_id").
		Select("accounts.name AS name, account_id, SUM(CASE WHEN type = 1 THEN value ELSE -value END) AS total").
		Where("transactions.team_id = ?", teamId).
		Where("account_id IS NOT NULL").
		Where("date <= ?", endDate).
		Group("accounts.name, account_id").
		Order("total desc").
		Scan(&accountBalances)

	var virtualAccountBalances []VirtualAccountBalance
	DB.Model(&models.Transaction{}).
		Select("account_virtual_id, SUM(CASE WHEN type = 1 THEN value ELSE -value END) as total").
		Where("team_id = ?", teamId).
		Where("account_virtual_id is not null").
		Where("date <= ?", endDate).
		Group("account_virtual_id").
		Scan(&virtualAccountBalances)

	// Soma dos saldos de contas reais
	var realTotal int
	for _, acc := range accountBalances {
		realTotal += acc.Total
	}

	// Soma dos saldos de contas virtuais
	var virtualTotal int
	for _, vAcc := range virtualAccountBalances {
		virtualTotal += vAcc.Total
	}

	// Balance = total das reais - total das virtuais
	finalBalance := realTotal - virtualTotal

	return AccountBalanceReport{
		Accounts:        accountBalances,
		VirtualAccounts: virtualAccountBalances,
		Balance:         finalBalance,
	}
}

func GetBalance(account models.Account, endDate string) int64 {
	type Balance struct {
		Total int64 `gorm:"column:total"`
	}
	var balance Balance
	query := DB.Model(&models.Transaction{}).
		Select("SUM(CASE WHEN type = 1 THEN value ELSE -value END) AS total").
		Where("date <= ?", endDate)

	if account.Virtual {
		query.Where("account_virtual_id = ?", account.ID)
	} else {
		query.Where("account_id = ?", account.ID)
	}

	query.Scan(&balance)

	return balance.Total
}

func GetExtractAllAccounts(startDate, endDate string, teamId uint) []models.Account {
	var accounts []models.Account
	DB.Model(&models.Account{}).
		Where("team_id = ?", teamId).
		Where("virtual = ?", false).
		Preload("RealTransactions", func(db *gorm.DB) *gorm.DB {
			return db.Where("date >= ?", startDate).
				Where("date <= ?", endDate).
				Order("date ASC")
		}).
		Find(&accounts)

	var virtualAccounts []models.Account
	DB.Model(&models.Account{}).
		Where("team_id = ?", teamId).
		Where("virtual = ?", true).
		Preload("VirtualTransactions", func(db *gorm.DB) *gorm.DB {
			return db.Where("date >= ?", startDate).
				Where("date <= ?", endDate).
				Order("date ASC")
		}).
		Find(&virtualAccounts)

	return append(accounts, virtualAccounts...)
}

func GetExtract(startDate, endDate string, teamId uint) []models.Account {
	var Accounts []models.Account
	DB.Model(&models.Account{}).
		Where("team_id = ?", teamId).
		Where("virtual = ?", true).
		Preload("VirtualTransactions", func(db *gorm.DB) *gorm.DB {
			return db.Where("date >= ?", startDate).Where("date <= ?", endDate)
		}).
		Find(&Accounts)

	return Accounts
}

func GetMonthlyMap(startDate, endDate string, teamId uint) []models.Category {
	var Categories []models.Category
	DB.Model(&models.Category{}).
		Where("team_id = ?", teamId).
		Where("use_map = ?", true).
		Order("name ASC").
		Preload("TransactionsMap", func(db *gorm.DB) *gorm.DB {
			return db.Where("date >= ?", startDate).
				Where("date <= ?", endDate).
				Where("is_transfer = ?", false).
				Order("date ASC")
		}).
		Find(&Categories)

	return Categories
}

func GetBalanceToMap(startDate, endDate string, teamId uint) (int64, int64) {
	type BalanceToMap struct {
		Total int64 `gorm:"column:total"`
	}

	var startBalanceMap BalanceToMap
	query := DB.Model(&models.Transaction{}).
		Select("SUM(CASE WHEN type = 1 THEN value ELSE -value END) AS total").
		Where("team_id = ?", teamId).
		Where("date <= ?", startDate).
		Where("account_id is not null")

	query.Scan(&startBalanceMap)

	var endBalanceMap BalanceToMap
	query2 := DB.Model(&models.Transaction{}).
		Select("SUM(CASE WHEN type = 1 THEN value ELSE -value END) AS total").
		Where("team_id = ?", teamId).
		Where("date <= ?", endDate).
		Where("account_id is not null")

	query2.Scan(&endBalanceMap)

	return startBalanceMap.Total, endBalanceMap.Total
}
