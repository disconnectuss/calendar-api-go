package common

import "strconv"

const (
	DefaultMaxResults int64 = 20
	MinMaxResults     int64 = 1
	MaxMaxResults     int64 = 100
)

func ParseMaxResults(raw string) int64 {
	if raw == "" {
		return DefaultMaxResults
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < MinMaxResults {
		return DefaultMaxResults
	}
	if n > MaxMaxResults {
		return MaxMaxResults
	}
	return n
}

type PaginatedResponse[T any] struct {
	Items         []T    `json:"items"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	ResultCount   int    `json:"resultCount"`
}

func NewPaginatedResponse[T any](items []T, nextPageToken string) PaginatedResponse[T] {
	if items == nil {
		items = []T{}
	}
	return PaginatedResponse[T]{
		Items:         items,
		NextPageToken: nextPageToken,
		ResultCount:   len(items),
	}
}
