package search

type FilterObject struct {
	Path  string `json:"path"`
	Operator string `json:"operator"`
	Values []string `json:"values"`
}

func RatingFilterTransfroms(rating string) (queryRating string) {
	switch rating {
	case "0":
		queryRating = "rating >= 0 AND rating <= 1"
	case "1":
		queryRating = "rating >= 1 AND rating <= 2"
	case "2":
		queryRating = "rating >= 2 AND rating <= `3"
	case "3":
		queryRating = "rating >= 3 AND rating <= 4"
	case "4":
		queryRating = "rating >= 4 AND rating <= 5"
	}
	return
}

func PriceFilterTransfroms(rating string) (queryPrice string) {
	switch rating {
	case "1000":
		queryPrice = "price.sellingPrice >= 0 AND price.sellingPrice <= 1000"
	case "5000":
		queryPrice = "price.sellingPrice >= 1000 AND price.sellingPrice <= 5000"
	case "10000":
		queryPrice = "price.sellingPrice >= 5000 AND price.sellingPrice <= 10000"
	case "10000above":
		queryPrice = "price.sellingPrice >= 10000"
	}
	return
}
