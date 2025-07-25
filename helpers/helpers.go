package helpers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func OnlyNumbers(input string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	numStrs := re.FindAllString(input, -1)

	if len(numStrs) == 0 {
		return 0, fmt.Errorf("nenhum número encontrado na string")
	}

	// Concatena todos os números encontrados
	var joinedNums strings.Builder
	for _, n := range numStrs {
		joinedNums.WriteString(n)
	}

	// Converte para int
	num, err := strconv.Atoi(joinedNums.String())
	if err != nil {
		return 0, fmt.Errorf("erro ao converter para inteiro: %w", err)
	}

	return num, nil
}

func DateToString(t time.Time) string {
	return t.Format("02/01/2006")
}

func StringToTime(date string) time.Time {
	timeValue, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Now()
	}

	return timeValue
}
