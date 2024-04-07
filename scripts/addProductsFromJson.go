package scripts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type productJSON struct {
	ID string `json:"_id" bson:"_id"`
}

func addProductsFromJson() error {

	jsonData, err := ioutil.ReadFile("haute_products.json")
	if err != nil {
		return err
	}

	// Unmarshal the byte slice into a slice of SKU structs
	var products []productJSON
	err = json.Unmarshal(jsonData, &products)
	if err != nil {
		return err
	}

	for idx, product := range products {
		fmt.Println("adding product: ", product.ID)
		err := AddOrReplaceProduct(product.ID)
		if err != nil {
			// panic(err)
			return err
		} else {
			fmt.Println("Added product ", idx)
		}
	}

	return nil
}
