package tmpl

import (
	"fmt"
	"strconv"
)

func TaxableValue(sellingPrice float64, tax string) float64 {
	if len(tax) == 0 {
		tax = "18"
	}

	taxPercentage, err := strconv.ParseFloat(tax, 64)
	if err != nil {
		fmt.Printf("error %s", err.Error())
		return 0
	}
	return sellingPrice / (1 + (taxPercentage)/100)
}

func TaxValue(sellingPrice float64, tax string) float64 {
	if len(tax) == 0 {
		tax = "18"
	}

	taxPercentage, err := strconv.ParseFloat(tax, 64)
	if err != nil {
		fmt.Printf("error %s", err.Error())
		return 0
	}
	taxableValue := sellingPrice / (1 + (taxPercentage)/100)
	return (taxableValue * taxPercentage) / 100
}
