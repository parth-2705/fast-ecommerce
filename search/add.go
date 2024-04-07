package search

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/tryamigo/themis"
)

func AddCatalogToIndex(products []interface{}) (err error) {
	
	defer sentry.Recover()
	
	params := [][]string{{"primaryKey", "id"}}
	body, _ := json.Marshal(products)
	status, resp, err := themis.HitAPIEndpoint2(SearchDBUrl+"/indexes/products/documents", http.MethodPost, body, SearchDBHeaders, params)
	if err != nil {
		fmt.Println(err)
		return
	} else if status >= 400 {
		fmt.Println(status, string(resp), err)
		return fmt.Errorf(fmt.Sprintln(status) + string(resp))
	}
	return
}

func UpdateProductByID(productID string) (err error) {

	defer sentry.Recover()

	params := [][]string{}
	product, err := models.GetCompleteProduct(productID)
	if err != nil {
		return err
	}
	products := []models.Product{product}
	body, err := json.Marshal(products)
	if err != nil {
		return err
	}
	status, resp, err := themis.HitAPIEndpoint2(SearchDBUrl+"/indexes/products/documents", http.MethodPost, body, SearchDBHeaders, params)
	if err != nil {
		fmt.Println(err)
		return
	} else if status >= 400 {
		fmt.Println(status, string(resp), err)
		return fmt.Errorf(fmt.Sprintln(status) + string(resp))
	}
	return
}
