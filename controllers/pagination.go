package controllers

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/search"
	utils "hermes/utils/queries"
	"math"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
)

func GetLimitFromQueryValue(limit string) int {
	limitInt := 20
	if limit != "" {
		limitInt, _ = strconv.Atoi(limit)
	}
	limitInt = int(math.Max(float64(limitInt), 20))
	return limitInt
}

func GetPageFromQueryValue(page string) int {
	pageInt := 1
	if page != "" {
		pageInt, _ = strconv.Atoi(page)
	}
	pageInt = int(math.Max(float64(pageInt), 1))
	return pageInt
}

type Pagination struct {
	Limit      int         `json:"limit,omitempty;"`
	Page       int         `json:"page,omitempty;"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Rows       interface{} `json:"rows"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return p.Page
}

func Paginate(collectionType string, pagination *Pagination, result interface{}, preSkip bson.A, postSkip bson.A) (*Pagination, error) {
	var count []map[string]interface{}

	countingSearchObject := preSkip
	countingSearchObject = append(countingSearchObject, bson.D{{Key: "$count", Value: "total_rows"}})
	countCur, err := db.StringToMongoCollectionMap[collectionType].Aggregate(context.Background(), countingSearchObject)
	if err != nil {
		return nil, err
	}
	err = countCur.All(context.Background(), &count)
	if err != nil {
		return nil, err
	}
	if len(count) == 0 {
		pagination.TotalPages = 1
		pagination.TotalRows = 0
		return pagination, nil
	}
	pagination.TotalRows = int64(count[0]["total_rows"].(int32))
	totalPages := int(math.Ceil(float64(pagination.TotalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages

	aggregateSearchObject := preSkip

	skip := (pagination.GetPage() - 1) * pagination.Limit
	limitQuery := bson.A{bson.D{{Key: "$limit", Value: skip + pagination.Limit}}, bson.D{{Key: "$skip", Value: skip}}}

	aggregateSearchObject = append(aggregateSearchObject, limitQuery...)
	aggregateSearchObject = append(aggregateSearchObject, postSkip...)

	cur, err := db.StringToMongoCollectionMap[collectionType].Aggregate(context.Background(), aggregateSearchObject)
	if err != nil {
		return pagination, err
	}

	err = cur.All(context.Background(), &result)
	if err != nil {
		return pagination, err
	}
	pagination.Rows = result
	return pagination, err
}

func ProductPaginate(pagination *Pagination, extraFilters ...bson.D) ([]models.Product, error) {
	var aggregateSearchObject bson.A
	var tempObject bson.A
	query := utils.ProductQuery
	aggregateSearchObject = query
	for _, filter := range extraFilters {
		aggregateSearchObject = append(aggregateSearchObject, filter)
		tempObject = append(tempObject, filter)
	}
	var count []map[string]interface{}
	countingSearchObject := tempObject
	countingSearchObject = append(countingSearchObject, bson.D{{Key: "$count", Value: "total_rows"}})
	countCur, err := db.ProductCollection.Aggregate(context.Background(), countingSearchObject)
	if err != nil {
		return nil, err
	}
	err = countCur.All(context.Background(), &count)
	if err != nil {
		return nil, err
	}
	if len(count) == 0 {
		pagination.TotalPages = 1
		pagination.TotalRows = 0
		return []models.Product{}, nil
	}
	pagination.TotalRows = int64(count[0]["total_rows"].(int32))
	totalPages := int(math.Ceil(float64(pagination.TotalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages
	skip := (pagination.GetPage() - 1) * pagination.Limit

	aggregateSearchObject = append(aggregateSearchObject, bson.D{{Key: "$skip", Value: skip}}, bson.D{{Key: "$limit", Value: pagination.Limit}})
	cur, err := db.ProductCollection.Aggregate(context.Background(), aggregateSearchObject)
	if err != nil {
		return nil, err
	}
	var temp []models.Product
	err = cur.All(context.Background(), &temp)
	return temp, err
}

func ProductPaginate2(pagination *Pagination, searchTerm string, filterRaw map[string]search.FilterObject, sort []search.SortObject) (products []models.Product, err error) {
	searchBody := map[string]interface{}{}
	if searchTerm != "" {
		searchBody["q"] = searchTerm
	}
	if len(filterRaw) > 0 {
		filterQuery := filterRawToFilterQuery(filterRaw)
		searchBody["filter"] = filterQuery
	}
	if len(sort) > 0 {
		sortTemp := []string{}
		for _, item := range sort {
			sortTemp = append(sortTemp, item.Path+":"+string(item.Order))
		}
		searchBody["sort"] = sortTemp
	} else {
		searchBody["sort"] = []string{"pageRanking:desc"}
	}
	searchBody["limit"] = pagination.Limit
	searchBody["page"] = pagination.Page
	products, totalResults, err := search.GetProducts(searchBody)
	pagination.TotalRows = int64(totalResults)
	totalPages := int(math.Ceil(float64(pagination.TotalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages
	return
}

func filterRawToFilterQuery(filterRaw map[string]search.FilterObject) (filterQuery string) {
	for key, obj := range filterRaw {
		value := obj.Values
		operator := obj.Operator
		path := obj.Path
		if key == "Price" {
			if len(value) > 1 {
				filterQuery += " AND ("
				for idx, item := range value {
					filterQuery += "(" + search.PriceFilterTransfroms(item) + ")"
					if idx != len(value)-1 {
						filterQuery += " OR "
					}
				}
				filterQuery += ")"
			} else {
				filterQuery += " AND " + search.PriceFilterTransfroms(value[0])
			}
		} else if key == "Rating" {
			if len(value) > 1 {
				filterQuery += " AND ("
				for idx, item := range value {
					filterQuery += "(" + search.RatingFilterTransfroms(item) + ")"
					if idx != len(value)-1 {
						filterQuery += " OR "
					}
				}
				filterQuery += ")"
			} else {
				filterQuery += " AND " + search.RatingFilterTransfroms(value[0])
			}
		} else {
			if len(value) > 1 {
				filterQuery += " AND ("
				for idx, item := range value {
					filterQuery += path + " " + operator + " " + item
					if idx != len(value)-1 {
						filterQuery += " OR "
					}
				}
				filterQuery += ")"
			} else {
				filterQuery += " AND " + path + " " + operator + " " + value[0]
			}
		}
	}
	fmt.Printf("filterQuery: %+v\n", filterQuery)
	filterQuery = filterQuery[5:]
	return
}
