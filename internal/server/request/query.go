package request

import (
	"net/http"
	"strconv"

	"github.com/aria3ppp/watchlist-server/internal/config"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SearchQuery struct {
	Query string `json:"query" query:"query"`
}

var _ validation.Validatable = SearchQuery{}

func (r SearchQuery) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Query,
			validation.Required,
			validation.Length(
				config.Config.Validation.Request.Search.Query.MinLength,
				config.Config.Validation.Request.Search.Query.MaxLength,
			),
		),
	)
}

func ParsePaginationQueries(req *http.Request) (page, perPage, offset int) {
	// set default values if query params not provided or mismatched
	page = parseIntDefault(
		req.URL.Query().Get(config.Config.Pagination.Page.VarName),
		config.Config.Pagination.Page.MinValue,
	)
	perPage = parseIntDefault(
		req.URL.Query().Get(config.Config.Pagination.PageSize.VarName),
		config.Config.Pagination.PageSize.DefaultValue,
	)
	// ensure page and per_page range of values are contraint to min and max values:
	// 		page >= min_value
	//		min_value <= per_page <= max_value
	if page < config.Config.Pagination.Page.MinValue {
		page = config.Config.Pagination.Page.MinValue
	}
	if perPage < config.Config.Pagination.PageSize.MinValue {
		perPage = config.Config.Pagination.PageSize.MinValue
	} else if perPage > config.Config.Pagination.PageSize.MaxValue {
		perPage = config.Config.Pagination.PageSize.MaxValue
	}
	return page, perPage, (page - 1) * perPage
}

func parseIntDefault(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return defaultValue
}
