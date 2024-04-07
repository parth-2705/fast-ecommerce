package search

import "os"

var SearchDBUrl string
var SearchDBAPIKey string
var SearchDBHeaders [][]string

func LoadSecrets() {
	SearchDBUrl = os.Getenv("SEARCH_DB_API_URL")
	SearchDBAPIKey = os.Getenv("SEARCH_DB_API_KEY")
	SearchDBHeaders = [][]string{{"Authorization", "Bearer " + SearchDBAPIKey}}
}
