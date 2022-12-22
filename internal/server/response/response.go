package response

type ResponseValue struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
	// response payload
	Payload any `json:"payload,omitempty"`
	// Page is 1-based and refers to the current page index/number.
	Page int `json:"page,omitempty"`
	// PageSize refers to the number of items on each page.
	PageSize int `json:"page_size,omitempty"`
	// PageCount is the number of all pages
	PageCount *int `json:"page_count,omitempty"`
	// TotalItems stands for the total number of items. If total is less than 0, it means total is unknown.
	TotalItems *int `json:"total_items,omitempty"`
}

func OK(payload any) *ResponseValue {
	return &ResponseValue{Status: StatusOK, Payload: &payload}
}

func Error(status Status, message ...string) *ResponseValue {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	}
	return &ResponseValue{Status: status, Message: msg}
}

func Paginated(
	page int,
	pageSize int,
	items any,
	totalItems int,
) *ResponseValue {
	pageCount := (totalItems + pageSize - 1) / pageSize
	return &ResponseValue{
		Status:     StatusOK,
		Page:       page,
		PageSize:   pageSize,
		PageCount:  &pageCount,
		Payload:    &items,
		TotalItems: &totalItems,
	}
}
