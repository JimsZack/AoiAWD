package storage

import (
	"context"
	"fmt"
)

type Storage interface {
	Save(ctx context.Context, collection, id string, doc interface{}) error
	Get(ctx context.Context, collection, id string) (interface{}, error)
	Paginate(ctx context.Context, collection string, page, count int) ([]interface{}, int64, error)
	Count(ctx context.Context, collection string) int64
	All(ctx context.Context, collection string) ([]interface{}, error)
	Close() error
}

func New(backend, path string) (Storage, error) {
	switch backend {
	case "memory", "":
		return NewMemory(), nil
	case "file", "json":
		return NewFile(path)
	default:
		return nil, fmt.Errorf("unsupported storage backend: %s (use memory or file)", backend)
	}
}

type PaginatedResult struct {
	Page     int           `json:"page"`
	LastPage int           `json:"last_page"`
	Data     []interface{} `json:"data"`
}

func NormalizePage(page, count, total int) (normalizedPage, offset, lastPage int) {
	if count <= 0 {
		count = 20
	}
	lastPage = (total + count - 1) / count
	if lastPage == 0 {
		lastPage = 1
	}
	if page <= 0 {
		page = lastPage
	}
	if page > lastPage {
		page = lastPage
	}
	offset = (page - 1) * count
	return page, offset, lastPage
}
