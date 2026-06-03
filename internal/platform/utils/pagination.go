package utils

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func NormalizePagination(limit, offset int) Pagination {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return Pagination{Limit: limit, Offset: offset}
}

