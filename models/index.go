package models

type IndexCreateable interface {
	CreateIndexes() error
}
