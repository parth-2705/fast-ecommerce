package db

// func CreateDummyData(n int) error {
// 	// Seed the random number generator
// 	rand.Seed(time.Now().UnixNano())

// 	// Generate n dummy products using gofakeit
// 	products := make([]models.Product, n)
// 	for i := 0; i < n; i++ {
// 		product := models.Product{
// 			ID:          gofakeit.UUID(),
// 			Name:        gofakeit.Name(),
// 			Description: gofakeit.Sentence(10),
// 			Price:       gofakeit.Price(10, 100),
// 			Quantity:    gofakeit.Number(1, 1000),
// 			Images:      make([]models.Image, 0),
// 		}
// 		for j := 0; j < gofakeit.Number(1, 5); j++ {
// 			image := models.Image{
// 				ID:  gofakeit.UUID(),
// 				Url: gofakeit.ImageURL(600, 400),
// 			}
// 			product.Images = append(product.Images, image)
// 		}
// 		products[i] = product
// 	}

// 	// Insert the products into the database
// 	for _, product := range products {
// 		_, err := ProductCollection.InsertOne(context.Background(), product)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
