package common

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
