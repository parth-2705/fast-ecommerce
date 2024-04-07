package search

type SortObject struct {
	Order SortOrder `json:"order"`
	Path  string 	`json:"path"`
}

type SortOrder string

const (
	ASC SortOrder = "asc"
	DESC SortOrder = "desc"
)