package search

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"net/http"

	"github.com/tryamigo/themis"
)

type SearchResponse struct {
	Hits               []models.Product `json:"hits"`
	Limit              int              `json:"limit"`
	Offset             int              `json:"offset"`
	EstimatedTotalHits int              `json:"estimatedTotalHits"`
	TotalHits          int              `json:"totalHits"`
}

func GetProducts(body map[string]interface{}) (products []models.Product, totalHits int, err error) {
	body["attributesToRetrieve"] = []string{"id", "name", "rating", "ratingCount", "price", "thumbnail", "brand"}
	bodyByte, _ := json.Marshal(body)
	status, resp, err := themis.HitAPIEndpoint2(SearchDBUrl+"/indexes/products/search", http.MethodPost, bodyByte, SearchDBHeaders, nil)
	if err != nil {
		fmt.Println(err)
		return
	} else if status >= 400 {
		fmt.Println(status, string(resp), err)
		return products, totalHits, fmt.Errorf(fmt.Sprintln(status) + string(resp))
	}
	var searchResp SearchResponse
	json.Unmarshal(resp, &searchResp)
	products = searchResp.Hits
	totalHits = searchResp.TotalHits
	if searchResp.TotalHits == 0 {
		totalHits = searchResp.EstimatedTotalHits
	}
	return
}
