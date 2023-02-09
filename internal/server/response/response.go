package response

type IDResponse struct {
	ID int `json:"id"`
}

func ID(id int) IDResponse {
	return IDResponse{ID: id}
}

type URIResponse struct {
	URI string `json:"uri"`
}

func URI(uri string) URIResponse {
	return URIResponse{URI: uri}
}

type PaginatedResponse[I any] struct {
	// response items
	Items []I `json:"items"`
	// Page is 0or1-based and refers to the current page index/number
	Page int `json:"page"`
	// PageSize refers to the number of items on each page
	PageSize int `json:"page_size"`
	// TotalPages is the number of all pages
	TotalPages int `json:"total_pages"`
	// TotalItems stands for the total number of items
	TotalItems int `json:"total_items"`
}

func Paginated[I any](
	page int,
	pageSize int,
	items []I,
	totalItems int,
) PaginatedResponse[I] {
	totalPages := (totalItems + pageSize - 1) / pageSize
	if page > totalPages {
		page = totalPages
	}
	return PaginatedResponse[I]{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalItems: totalItems,
	}
}
