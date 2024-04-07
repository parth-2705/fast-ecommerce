package scripts

import (
	"fmt"
	"hermes/configs/Redis"
	"hermes/models"
)

func imagesToMedia() (err error) {
	products, err := models.GetAllProducts()
	if err != nil {
		return
	}
	totalProducts := len(products)
	failedProducts, passedProducts := 0, 0
	for _, product := range products {
		media := []models.MediaObject{}
		for _, image := range product.Images {
			temp := models.MediaObject{
				ID:   string(image),
				Type: models.TypeImage,
			}
			media = append(media, temp)
		}
		product.Media = media
		err = product.Update()
		if err != nil {
			failedProducts++
			continue
		}
		err = Redis.DeleteProductCacheByID(product.ID)
		if err != nil {
			failedProducts++
			continue
		}
		passedProducts++
	}
	fmt.Printf("totalProducts: %v\n", totalProducts)
	fmt.Printf("passedProducts: %v\n", passedProducts)
	fmt.Printf("failedProducts: %v\n", failedProducts)

	variants, err := models.GetAllVariants()
	if err != nil {
		return
	}
	totalVariants := len(variants)
	failedVariants, passedVariants := 0, 0
	for _, variant := range variants {
		media := []models.MediaObject{}
		for _, image := range variant.Images {
			temp := models.MediaObject{
				ID:   string(image),
				Type: models.TypeImage,
			}
			media = append(media, temp)
		}
		variant.Media = media
		err = variant.Update()
		if err != nil {
			failedVariants++
			continue
		}
		passedVariants++
	}
	fmt.Printf("totalVariants: %v\n", totalVariants)
	fmt.Printf("passedVariants: %v\n", passedVariants)
	fmt.Printf("failedVariants: %v\n", failedVariants)
	return
}
