package helpers

import (
	"fmt"
	"regexp"
	"strconv"
)

func OnlyNumbers(input string) (float64, error) {
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindAllString(input, -1)

	if len(numStr) == 0 {
		return 0, fmt.Errorf("nenhum número encontrado na string")
	}

	// Concatenar todos os números encontrados
	joinedNums := ""
	for _, n := range numStr {
		joinedNums += n
	}

	// Converter para float64
	num, err := strconv.ParseFloat(joinedNums, 64)
	if err != nil {
		return 0, err
	}

	return num / 100, nil
}
