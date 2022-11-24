package response

type ResponseValue struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
	// response payload
	Payload *any `json:"payload,omitempty"`
	// Page is 1-based and refers to the current page index/number.
	Page int `json:"page,omitempty"`
	// PerPage refers to the number of items on each page.
	PerPage int `json:"per_page,omitempty"`
	// PageCount is the number of all pages
	PageCount *int `json:"page_count,omitempty"`
	// TotalItems stands for the total number of items. If total is less than 0, it means total is unknown.
	TotalItems *int `json:"total_items,omitempty"`
}

// // http StatusCreated 201
// func Success(payload any, message ...string) *ResponseValue {
// 	var msg string
// 	if len(message) > 0 {
// 		msg = message[0]
// 	}
// 	return &ResponseValue{Status: StatusSuccess, Message: msg, Payload: &payload}
// }

func OK(payload any, message ...string) *ResponseValue {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	}
	return &ResponseValue{Status: StatusOK, Message: msg, Payload: &payload}
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
	perPage int,
	items any,
	totalItems int,
	message ...string,
) *ResponseValue {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	}
	pageCount := (totalItems + perPage - 1) / perPage
	return &ResponseValue{
		Status:     StatusOK,
		Message:    msg,
		Page:       page,
		PerPage:    perPage,
		PageCount:  &pageCount,
		Payload:    &items,
		TotalItems: &totalItems,
	}
}
