package tmpl

import (
	"fmt"
	"math"
	"strconv"

	"github.com/brianvoe/gofakeit/v6"
)

// round off to n decimal places
func Round(x float64, n int) float64 {
	p := math.Pow(10, float64(n))
	return math.Round(x*p) / p
}

// remove decimal places and return an integer
func RoundToInt(x float64) int {
	return int(math.Round(x))
}

func Minus(x int, y int) int {
	return x - y
}

func MinusInt(x int, y int) int {
	return x - y
}

func Add(x float64, y float64) float64 {
	return x + y
}

func AddInt(x int, y int) int {
	return x + y
}

// calculate volumetric weight
func VolumetricWeight(variant map[string]interface{}) string {
	var length, breadth, height float64
	l, ok := variant["length"].(string)
	if ok {
		length, _ = strconv.ParseFloat(l, 64)
	}
	b, ok := variant["breadth"].(string)
	if ok {
		breadth, _ = strconv.ParseFloat(b, 64)
	}
	h, ok := variant["height"].(string)
	if ok {
		height, _ = strconv.ParseFloat(h, 64)
	}

	volumetricWeight := (length * breadth * height) / 5000

	return fmt.Sprint(volumetricWeight)
}

// calculate discount percenta
func CalculateDiscount(sellingPrice float64, mrp float64) float64 {
	return math.Round(((mrp - sellingPrice) / mrp) * 100)
}

func RandomInt() int {
	return gofakeit.IntRange(118, 137)
}
