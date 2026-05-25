package utils

// Pagination holds offset and limit parameters.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// NormalizePagination returns normalized limit and offset values.
func NormalizePagination(limit, offset int) Pagination {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return Pagination{Limit: limit, Offset: offset}
}
