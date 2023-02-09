package request_test

import (
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/aria3ppp/watchlist-server/internal/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/require"
)

func TestPaginationSortingQuery_Validate(t *testing.T) {
	nonExistentField := "non-existent-field"
	testCases := []struct {
		name     string
		query    request.PaginationSortingQuery
		model    string
		expError error
	}{
		{
			name:     "tc1",
			query:    request.PaginationSortingQuery{},
			expError: nil,
		},
		{
			name: "tc2",
			query: request.PaginationSortingQuery{
				PaginationQuery: request.PaginationQuery{
					Page:     -1,
					PageSize: -1,
				},
			},
			expError: validation.Errors{
				"page": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.Page.MinValue,
					},
				),
				"page_size": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.PageSize.MinValue,
					},
				),
			},
		},
		{
			name: "tc3",
			query: request.PaginationSortingQuery{
				PaginationQuery: request.PaginationQuery{
					PageSize: config.Config.Validation.Pagination.PageSize.MaxValue + 1,
				},
			},
			expError: validation.Errors{
				"page_size": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.PageSize.MaxValue,
					},
				),
			},
		},
		{
			name: "tc4",
			query: request.PaginationSortingQuery{
				SortingQuery: request.SortingQuery{
					SortField: models.SeriesColumns.ID,
					SortOrderQuery: request.SortOrderQuery{
						SortOrder: request.SortOrderAsc,
					},
				},
			},
			model:    models.TableNames.Serieses,
			expError: nil,
		},
		{
			name: "tc5",
			query: request.PaginationSortingQuery{
				SortingQuery: request.SortingQuery{
					SortField: models.FilmColumns.ContributedAt,
					SortOrderQuery: request.SortOrderQuery{
						SortOrder: request.SortOrderDesc,
					},
				},
			},
			model:    models.TableNames.Films,
			expError: nil,
		},
		{
			name: "tc6",
			query: request.PaginationSortingQuery{
				SortingQuery: request.SortingQuery{
					SortField: nonExistentField,
					SortOrderQuery: request.SortOrderQuery{
						SortOrder: "invalid_sort_order",
					},
				},
			},
			model: models.TableNames.Films,
			expError: validation.Errors{
				"sort_field": validator.ErrInvalidFieldOfModel.SetParams(
					map[string]any{
						"field": nonExistentField,
						"model": models.TableNames.Films,
					},
				),
				"sort_order": validation.ErrInInvalid,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(
				tc.expError,
				tc.query.SetValidationModel(tc.model).Validate(),
			)
		})
	}
}

func TestPaginationSortOrderQuery_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		query    request.PaginationSortOrderQuery
		expError error
	}{
		{
			name:     "tc1",
			query:    request.PaginationSortOrderQuery{},
			expError: nil,
		},
		{
			name: "tc2",
			query: request.PaginationSortOrderQuery{
				PaginationQuery: request.PaginationQuery{
					Page:     -1,
					PageSize: -1,
				},
			},
			expError: validation.Errors{
				"page": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.Page.MinValue,
					},
				),
				"page_size": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.PageSize.MinValue,
					},
				),
			},
		},
		{
			name: "tc3",
			query: request.PaginationSortOrderQuery{
				PaginationQuery: request.PaginationQuery{
					PageSize: config.Config.Validation.Pagination.PageSize.MaxValue + 1,
				},
			},
			expError: validation.Errors{
				"page_size": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.PageSize.MaxValue,
					},
				),
			},
		},
		{
			name: "tc4",
			query: request.PaginationSortOrderQuery{
				SortOrderQuery: request.SortOrderQuery{
					SortOrder: request.SortOrderAsc,
				},
			},
			expError: nil,
		},
		{
			name: "tc5",
			query: request.PaginationSortOrderQuery{
				SortOrderQuery: request.SortOrderQuery{
					SortOrder: request.SortOrderDesc,
				},
			},
			expError: nil,
		},
		{
			name: "tc6",
			query: request.PaginationSortOrderQuery{
				SortOrderQuery: request.SortOrderQuery{
					SortOrder: "invalid_sort_order",
				},
			},
			expError: validation.Errors{
				"sort_order": validation.ErrInInvalid,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.query.Validate())
		})
	}
}

func TestSearchPaginationQuery_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		query    request.SearchPaginationQuery
		expError error
	}{
		{
			name:  "tc1",
			query: request.SearchPaginationQuery{},
			expError: validation.Errors{
				"query": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			query: request.SearchPaginationQuery{
				Query: "q",
			},
			expError: validation.Errors{
				"query": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Request.Search.Query.MinLength,
						"max": config.Config.Validation.Request.Search.Query.MaxLength,
					},
				),
			},
		},
		{
			name: "tc3",
			query: request.SearchPaginationQuery{
				Query: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.Request.Search.Query.MaxLength,
				),
			},
			expError: validation.Errors{
				"query": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Request.Search.Query.MinLength,
						"max": config.Config.Validation.Request.Search.Query.MaxLength,
					},
				),
			},
		},
		{
			name: "tc4",
			query: request.SearchPaginationQuery{
				Query: "query",
				PaginationQuery: request.PaginationQuery{
					Page:     -1,
					PageSize: -1,
				},
			},
			expError: validation.Errors{
				"page": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.Page.MinValue,
					},
				),
				"page_size": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.PageSize.MinValue,
					},
				),
			},
		},
		{
			name: "tc5",
			query: request.SearchPaginationQuery{
				Query: "query",
				PaginationQuery: request.PaginationQuery{
					PageSize: config.Config.Validation.Pagination.PageSize.MaxValue + 1,
				},
			},
			expError: validation.Errors{
				"page_size": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.PageSize.MaxValue,
					},
				),
			},
		},
		{
			name: "tc6",
			query: request.SearchPaginationQuery{
				Query: "query",
				PaginationQuery: request.PaginationQuery{
					Page:     config.Config.Validation.Pagination.Page.MinValue,
					PageSize: config.Config.Validation.Pagination.PageSize.MinValue,
				},
			},
			expError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.query.Validate())
		})
	}
}

func TestWatchlistGetQuery_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		query    request.WatchlistGetQuery
		expError error
	}{
		{
			name:     "tc1",
			query:    request.WatchlistGetQuery{},
			expError: nil,
		},
		{
			name: "tc2",
			query: request.WatchlistGetQuery{
				PaginationQuery: request.PaginationQuery{
					Page:     -1,
					PageSize: -1,
				},
			},
			expError: validation.Errors{
				"page": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.Page.MinValue,
					},
				),
				"page_size": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.PageSize.MinValue,
					},
				),
			},
		},
		{
			name: "tc3",
			query: request.WatchlistGetQuery{
				PaginationQuery: request.PaginationQuery{
					PageSize: config.Config.Validation.Pagination.PageSize.MaxValue + 1,
				},
			},
			expError: validation.Errors{
				"page_size": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Pagination.PageSize.MaxValue,
					},
				),
			},
		},
		{
			name: "tc4",
			query: request.WatchlistGetQuery{
				SortOrderQuery: request.SortOrderQuery{
					SortOrder: request.SortOrderAsc,
				},
			},
			expError: nil,
		},
		{
			name: "tc5",
			query: request.WatchlistGetQuery{
				SortOrderQuery: request.SortOrderQuery{
					SortOrder: request.SortOrderDesc,
				},
			},
			expError: nil,
		},
		{
			name: "tc6",
			query: request.WatchlistGetQuery{
				SortOrderQuery: request.SortOrderQuery{
					SortOrder: "invalid_sort_order",
				},
			},
			expError: validation.Errors{
				"sort_order": validation.ErrInInvalid,
			},
		},
		{
			name: "tc7",
			query: request.WatchlistGetQuery{
				Filter: request.WatchlistFilterWatched,
			},
			expError: nil,
		},
		{
			name: "tc8",
			query: request.WatchlistGetQuery{
				Filter: request.WatchlistFilterNotWatched,
			},
			expError: nil,
		},
		{
			name: "tc9",
			query: request.WatchlistGetQuery{
				Filter: request.WatchlistFilterAll,
			},
			expError: nil,
		},
		{
			name: "tc10",
			query: request.WatchlistGetQuery{
				Filter: "invalid_items",
			},
			expError: validation.Errors{
				"filter": validation.ErrInInvalid,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.query.Validate())
		})
	}
}

func TestWatchlistAddQuery_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		params   request.WatchlistAddQuery
		expError error
	}{
		{
			name:   "tc1",
			params: request.WatchlistAddQuery{},
			expError: validation.Errors{
				"film_id": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			params: request.WatchlistAddQuery{
				FilmID: -1,
			},
			expError: validation.Errors{
				"film_id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			},
		},
		{
			name: "tc3",
			params: request.WatchlistAddQuery{
				FilmID: 1,
			},
			expError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.params.Validate())
		})
	}
}
