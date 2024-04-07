package scripts

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"hermes/search"
	"net/http"

	"github.com/tryamigo/themis"
)

func DeleteAllDocuments() (err error) {
	status, resp, err := themis.HitAPIEndpoint2(search.SearchDBUrl+"/indexes/products/documents", http.MethodDelete, nil, search.SearchDBHeaders, nil)
	if err != nil {
		fmt.Println(err)
		return
	} else if status >= 400 {
		fmt.Println(status, resp, err)
		return fmt.Errorf(fmt.Sprintln(status) + string(resp))
	}
	fmt.Println("DELETED")
	fmt.Println(string(resp))
	return
}

func DeleteDocumentFromSearchDB(ID string) (err error) {
	status, resp, err := themis.HitAPIEndpoint2(search.SearchDBUrl+"/indexes/products/documents/"+ID, http.MethodDelete, nil, search.SearchDBHeaders, nil)
	if err != nil {
		fmt.Println(err)
		return
	} else if status >= 400 {
		fmt.Println(status, resp, err)
		return fmt.Errorf(fmt.Sprintln(status) + string(resp))
	}
	fmt.Println("DELETED")
	fmt.Println(string(resp))
	return
}

func AddCatalogToIndex() (err error) {
	params := [][]string{{"primaryKey", "id"}}
	products, err := models.GetFullProducts()
	fmt.Println("total products", len(products), search.SearchDBHeaders)
	body, _ := json.Marshal(products)
	status, resp, err := themis.HitAPIEndpoint2(search.SearchDBUrl+"/indexes/products/documents", http.MethodPost, body, search.SearchDBHeaders, params)
	if err != nil {
		fmt.Println(err)
		return
	} else if status >= 400 {
		fmt.Println(status, string(resp), err)
		return fmt.Errorf(fmt.Sprintln(status) + string(resp))
	}
	fmt.Println("ADDED")
	fmt.Println(string(resp))
	return
}

func AddOrReplaceProduct(id string) (err error) {
	params := [][]string{}
	product, err := models.GetCompleteProduct(id)
	if err != nil {
		return err
	}
	products := []models.Product{product}
	body, err := json.Marshal(products)
	if err != nil {
		return err
	}
	status, resp, err := themis.HitAPIEndpoint2(search.SearchDBUrl+"/indexes/products/documents", http.MethodPost, body, search.SearchDBHeaders, params)
	if err != nil {
		fmt.Println(err)
		return
	} else if status >= 400 {
		fmt.Println(status, string(resp), err)
		return fmt.Errorf(fmt.Sprintln(status) + string(resp))
	}

	return
}
