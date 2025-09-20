package ofx

import (
	"bufio"
	"fmt"
	"jc-financas/helpers"
	"jc-financas/models"
	"mime/multipart"
	"regexp"
	"strings"
	"time"
)

func TranslaterOfxInter(file *multipart.FileHeader, account models.Account) ([]models.Transaction, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer src.Close()

	scanner := bufio.NewScanner(src)

	type TransactionOFX struct {
		Type        string
		Date        string
		Amount      string
		FITID       string
		Description string
	}

	var transactions []TransactionOFX
	var current TransactionOFX
	rgxDate := regexp.MustCompile(`\d{8}`)

	fmt.Println("---- Lendo OFX do Inter ----")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fmt.Println("Linha:", line)

		switch {
		case strings.HasPrefix(line, "<STMTTRN>"):
			current = TransactionOFX{}

		case strings.HasPrefix(line, "<TRNTYPE>"):
			current.Type = strings.TrimPrefix(line, "<TRNTYPE>")

		case strings.HasPrefix(line, "<DTPOSTED>"):
			match := rgxDate.FindString(line)
			if match != "" {
				parsedDate, _ := time.Parse("20060102", match)
				current.Date = parsedDate.Format("2006-01-02")
			}

		case strings.HasPrefix(line, "<TRNAMT>"):
			current.Amount = strings.TrimPrefix(line, "<TRNAMT>")

		case strings.HasPrefix(line, "<FITID>"):
			current.FITID = cleanOFXTag(line, "FITID")

		case strings.HasPrefix(line, "<MEMO>"):
			current.Description = cleanOFXTag(line, "MEMO")

		case strings.HasPrefix(line, "</STMTTRN>"):
			transactions = append(transactions, current)
			fmt.Println(">> Transação capturada:", current)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo: %v", err)
	}

	fmt.Println("---- Resultado parse OFX ----")
	fmt.Printf("%-8s %-12s %-10s %-20s %s\n", "Tipo", "Data", "Valor", "Código", "Descrição")
	fmt.Println(strings.Repeat("-", 80))

	var result []models.Transaction
	for _, tx := range transactions {
		fmt.Printf("%-8s %-12s %-10s %-20s %s\n", tx.Type, tx.Date, tx.Amount, tx.FITID, tx.Description)

		record := models.Transaction{}
		record.TeamID = account.TeamID
		record.AccountID = &account.ID
		record.Date = tx.Date
		record.Description = tx.Description
		record.ExternalId = &tx.FITID

		// Converte valor para int (ex: -200.00 => -20000)
		valor, err := helpers.OnlyNumbers(tx.Amount)
		if err != nil {
			return nil, fmt.Errorf("erro ao interpretar valor: %v", err)
		}
		record.Value = valor

		// Determina se é entrada ou saída
		record.Type = 1
		if strings.HasPrefix(tx.Amount, "-") {
			record.Type = 2
		}

		result = append(result, record)
	}

	return result, nil
}

func cleanOFXTag(line, tag string) string {
	value := strings.TrimPrefix(line, "<"+tag+">")
	value = strings.TrimSuffix(value, "</"+tag+">")
	return strings.TrimSpace(value)
}
