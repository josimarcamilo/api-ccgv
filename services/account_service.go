package services

import (
	"jc-financas/consts"
	"jc-financas/helpers"
	"jc-financas/models"
	"jc-financas/repositories"
)

type ReportExtractAllAccount struct {
	AccountId    uint
	AccountName  string
	StartDate    string
	EndDate      string
	StartBalance int64
	EndBalance   int64
	TotalEntry   int64
	TotalExit    int64
	Transactions []models.Transaction
}

type ReportMonthlyMap struct {
	StartDate    string
	EndDate      string
	StartBalance int64
	TotalEntry   int64
	TotalExit    int64
	EndBalance   int64
	Totals       map[uint]int64
	Repasse      []ReportTipoRepasse
	Entry        []ReportCategoryMap
	Exit         []ReportCategoryMap
}

type ReportCategoryMap struct {
	ID           uint
	Name         string
	TipoRepasse  int8
	Transactions []models.Transaction
}

type ReportTipoRepasse struct {
	ID               uint
	Description      string
	TipoRepasse      int8
	CalculationBasis int64
	Repasse          int64
}

func GetExtractAllAccounts(startDate, endDate string, teamId uint) []ReportExtractAllAccount {
	var accounts []models.Account
	var result []ReportExtractAllAccount
	var saldoIncial int64
	var saldoFinal int64
	accounts = repositories.GetExtractAllAccounts(startDate, endDate, teamId)
	// calcular saldo inicial
	// calcular saldo final
	// somar entradas de Transactions
	// somar saidas de Transactions
	for _, account := range accounts {
		saldoIncial = 0
		diaAnterior := helpers.StringToTime(startDate).AddDate(0, 0, -1).Format("2006-01-02")
		saldoIncial = repositories.GetBalance(account, diaAnterior)
		saldoFinal = 0
		saldoFinal = repositories.GetBalance(account, endDate)
		var totalEntradas int64
		totalEntradas = 0
		var totalSaidas int64
		totalEntradas = 0
		transactions := account.RealTransactions
		if account.Virtual {
			transactions = account.VirtualTransactions
		}
		for _, tran := range transactions {
			if tran.IsTransfer {
				continue
			}
			if tran.Type == consts.TransactionTypeEntrada {
				totalEntradas = totalEntradas + int64(tran.Value)
			}
			if tran.Type == consts.TransactionTypeSaida {
				totalSaidas = totalSaidas + int64(tran.Value)
			}
		}
		result = append(result, ReportExtractAllAccount{
			AccountId:    account.ID,
			AccountName:  account.Name,
			StartDate:    startDate,
			EndDate:      endDate,
			StartBalance: saldoIncial,
			EndBalance:   saldoFinal,
			TotalEntry:   totalEntradas,
			TotalExit:    totalSaidas,
			Transactions: transactions,
		})
	}

	return result
}

func GetMonthlyMap(startDate, endDate string, teamId uint) ReportMonthlyMap {
	var categoriesMap []models.Category
	var entryCategories []ReportCategoryMap
	var exitCategories []ReportCategoryMap
	var categoriesTotals = make(map[uint]int64)

	totalEntradas := int64(0)
	totalSaidas := int64(0)
	somaRepasse2_5 := int64(0)
	somaRepasse10 := int64(0)
	somaRepasse75 := int64(0)

	saldoIncial, saldoFinal := repositories.GetBalanceToMap(startDate, endDate, teamId)
	categoriesMap = repositories.GetMonthlyMap(startDate, endDate, teamId)

	for _, category := range categoriesMap {
		if category.Type == consts.CategoryTypeEntry {
			entryCategories = append(entryCategories, ReportCategoryMap{
				ID:           category.ID,
				Name:         category.Name,
				TipoRepasse:  category.TipoRepasse,
				Transactions: category.TransactionsMap,
			})
		}

		if category.Type == consts.CategoryTypeExit {
			exitCategories = append(exitCategories, ReportCategoryMap{
				ID:           category.ID,
				Name:         category.Name,
				TipoRepasse:  category.TipoRepasse,
				Transactions: category.TransactionsMap,
			})
		}

		sumCategory := int64(0)
		for _, tran := range category.TransactionsMap {
			if category.TipoRepasse == consts.TipoRepasse2_5 {
				somaRepasse2_5 = somaRepasse2_5 + int64(tran.Value)
			}
			if category.TipoRepasse == consts.TipoRepasse10 {
				somaRepasse10 = somaRepasse10 + int64(tran.Value)
			}
			if category.TipoRepasse == consts.TipoRepasse75 {
				somaRepasse75 = somaRepasse75 + int64(tran.Value)
			}
			if tran.Type == consts.TransactionTypeEntrada {
				totalEntradas = totalEntradas + int64(tran.Value)
			}
			if tran.Type == consts.TransactionTypeSaida {
				totalSaidas = totalSaidas + int64(tran.Value)
			}
			sumCategory = sumCategory + int64(tran.Value)

		}
		categoriesTotals[category.ID] = sumCategory
	}

	return ReportMonthlyMap{
		StartDate:    startDate,
		EndDate:      endDate,
		StartBalance: saldoIncial,
		EndBalance:   saldoFinal,
		TotalEntry:   totalEntradas,
		TotalExit:    totalSaidas,
		Totals:       categoriesTotals,
		Repasse: []ReportTipoRepasse{
			{
				ID:               1,
				Description:      "2,5%",
				TipoRepasse:      consts.TipoRepasse2_5,
				CalculationBasis: somaRepasse2_5,
				Repasse:          (somaRepasse2_5 * 25) / 1000,
			},
			{
				ID:               2,
				Description:      "10%",
				TipoRepasse:      consts.TipoRepasse10,
				CalculationBasis: somaRepasse10,
				Repasse:          (somaRepasse10 * 10) / 100,
			},
			{
				ID:               3,
				Description:      "75%",
				TipoRepasse:      consts.TipoRepasse75,
				CalculationBasis: somaRepasse75,
				Repasse:          (somaRepasse75 * 75) / 100,
			},
		},
		Entry: entryCategories,
		Exit:  exitCategories,
	}
}
