package request

import (
	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PaginationSortingQuery struct {
	PaginationQuery
	SortingQuery
}

func (q *PaginationSortingQuery) SetValidationModel(
	model string,
) *paginationSortingQueryValidate {
	q.SortingQuery.model = model
	return (*paginationSortingQueryValidate)(q)
}

type paginationSortingQueryValidate PaginationSortingQuery

var _ validation.Validatable = paginationSortingQueryValidate{}

func (r paginationSortingQueryValidate) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(&r.PaginationQuery),
		validation.Field(&r.SortingQuery),
	)
}

func (q *PaginationSortingQuery) SetQueryIfNotSet(
	alt PaginationSortingQuery,
) paginationSortingQueryToQueryOptions {
	if q.Page == 0 {
		q.Page = alt.Page
	}
	if q.PageSize == 0 {
		q.PageSize = alt.PageSize
	}
	if q.SortField == "" {
		q.SortField = alt.SortField
	}
	if q.SortOrder == "" {
		q.SortOrder = alt.SortOrder
	}

	return paginationSortingQueryToQueryOptions(*q)
}

type paginationSortingQueryToQueryOptions PaginationSortingQuery

func (q paginationSortingQueryToQueryOptions) ToQueryOptions() query.Options {
	return query.Options{
		Offset:    q.PaginationQuery.Offset(),
		Limit:     q.PaginationQuery.Limit(),
		SortField: q.SortField,
		SortOrder: q.SortOrder,
	}
}

////////////////////////////////////////////////////////////////////////////////

type PaginationSortOrderQuery struct {
	PaginationQuery
	SortOrderQuery
}

var _ validation.Validatable = PaginationSortOrderQuery{}

func (r PaginationSortOrderQuery) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(&r.PaginationQuery),
		validation.Field(&r.SortOrderQuery),
	)
}

func (q *PaginationSortOrderQuery) SetQueryIfNotSet(
	alt PaginationSortOrderQuery,
) paginationSortOrderQueryToQueryOptions {
	if q.Page == 0 {
		q.Page = alt.Page
	}
	if q.PageSize == 0 {
		q.PageSize = alt.PageSize
	}
	if q.SortOrder == "" {
		q.SortOrder = alt.SortOrder
	}

	return paginationSortOrderQueryToQueryOptions(*q)
}

type paginationSortOrderQueryToQueryOptions PaginationSortOrderQuery

func (q paginationSortOrderQueryToQueryOptions) ToQueryOptions() query.SortOrderOptions {
	return query.SortOrderOptions{
		Offset:    q.PaginationQuery.Offset(),
		Limit:     q.PaginationQuery.Limit(),
		SortOrder: q.SortOrder,
	}
}

////////////////////////////////////////////////////////////////////////////////

type SearchPaginationQuery struct {
	Query string `query:"query" url:"query" json:"query"`
	PaginationQuery
}

var _ validation.Validatable = SearchPaginationQuery{}

func (r SearchPaginationQuery) Validate() error {
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
		validation.Field(&r.PaginationQuery),
	)
}

func (q *SearchPaginationQuery) SetQueryIfNotSet(
	alt PaginationQuery,
) searchPaginationQueryToQueryOptions {
	if q.Page == 0 {
		q.Page = alt.Page
	}
	if q.PageSize == 0 {
		q.PageSize = alt.PageSize
	}
	return searchPaginationQueryToQueryOptions(*q)
}

type searchPaginationQueryToQueryOptions SearchPaginationQuery

func (q searchPaginationQueryToQueryOptions) ToQueryOptions() query.SearchOptions {
	return query.SearchOptions{
		Query: q.Query,
		From:  q.PaginationQuery.Offset(),
		Size:  q.PaginationQuery.Limit(),
	}
}

////////////////////////////////////////////////////////////////////////////////

type PaginationQuery struct {
	Page     int `query:"page"      url:"page"      json:"page"`
	PageSize int `query:"page_size" url:"page_size" json:"page_size"`
}

var _ validation.Validatable = PaginationQuery{}

func (r PaginationQuery) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Page,
			validation.Min(config.Config.Validation.Pagination.Page.MinValue),
		),
		validation.Field(
			&r.PageSize,
			validation.Min(
				config.Config.Validation.Pagination.PageSize.MinValue,
			),
			validation.Max(
				config.Config.Validation.Pagination.PageSize.MaxValue,
			),
		),
	)
}

func (p PaginationQuery) Offset() int {
	return p.PageSize * (p.Page - config.Config.Validation.Pagination.Page.MinValue)
}

func (p PaginationQuery) Limit() int {
	return p.PageSize
}

////////////////////////////////////////////////////////////////////////////////

type SortingQuery struct {
	SortField string `query:"sort_field" url:"sort_field" json:"sort_field"`
	SortOrderQuery

	model string `query:"-" url:"-" json:"-"`
}

var _ validation.Validatable = SortingQuery{}

func (r SortingQuery) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.SortField,
			validator.IsValidFieldOfModel(r.model),
		),
		validation.Field(&r.SortOrderQuery),
	)
}

const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

type SortOrderQuery struct {
	SortOrder string `query:"sort_order" url:"sort_order" json:"sort_order"`
}

var _ validation.Validatable = SortOrderQuery{}

func (r SortOrderQuery) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.SortOrder,
			validation.In(SortOrderAsc, SortOrderDesc),
		),
	)
}

////////////////////////////////////////////////////////////////////////////////

const (
	WatchlistFilterWatched    = "watched"
	WatchlistFilterNotWatched = "not-watched"
	WatchlistFilterAll        = "all"
)

type WatchlistGetQuery struct {
	PaginationQuery
	SortOrderQuery
	Filter string `query:"filter" url:"filter" json:"filter"`
}

var _ validation.Validatable = WatchlistGetQuery{}

func (r WatchlistGetQuery) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(&r.PaginationQuery),
		validation.Field(&r.SortOrderQuery),
		validation.Field(
			&r.Filter,
			validation.In(
				WatchlistFilterWatched,
				WatchlistFilterNotWatched,
				WatchlistFilterAll,
			),
		),
	)
}

func (q *WatchlistGetQuery) SetQueryIfNotSet(
	alt WatchlistGetQuery,
) watchlistGetQueryToQueryOptions {
	if q.Page == 0 {
		q.Page = alt.Page
	}
	if q.PageSize == 0 {
		q.PageSize = alt.PageSize
	}
	if q.SortOrder == "" {
		q.SortOrder = alt.SortOrder
	}
	if q.Filter == "" {
		q.Filter = alt.Filter
	}

	return watchlistGetQueryToQueryOptions(*q)
}

type watchlistGetQueryToQueryOptions WatchlistGetQuery

func (q watchlistGetQueryToQueryOptions) ToQueryOptions() query.WatchlistOptions {
	var whereTimeWatched string
	if q.Filter == WatchlistFilterWatched {
		whereTimeWatched = repo.RawSqlWhereTimeWatchedIsNotNull
	} else if q.Filter == WatchlistFilterNotWatched {
		whereTimeWatched = repo.RawSqlWhereTimeWatchedIsNull
	} else if q.Filter == WatchlistFilterAll {
		whereTimeWatched = repo.RawSqlWhereTimeWatchedEmptyClause
	}
	return query.WatchlistOptions{
		Offset:           q.PaginationQuery.Offset(),
		Limit:            q.PaginationQuery.Limit(),
		SortOrder:        q.SortOrder,
		WhereTimeWatched: whereTimeWatched,
	}
}

////////////////////////////////////////////////////////////////////////////////

type WatchlistAddQuery struct {
	FilmID int `query:"film_id" url:"film_id" json:"film_id"`
}

var _ validation.Validatable = WatchlistAddQuery{}

func (r WatchlistAddQuery) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.FilmID,
			validation.Required,
			validation.Min(1),
		),
	)
}
