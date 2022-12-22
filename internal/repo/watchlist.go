package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/watchlist"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
)

func (repo *Repository) WatchlistGet(
	ctx context.Context,
	userID int,
	queryOptions query.WatchlistOptions,
) (watchlist []*watchlist.Item, err error) {
	// query
	rows, err := repo.exec.QueryContext(
		ctx,
		fmt.Sprintf(
			watchfilmGetAllQuery,
			queryOptions.WhereTimeWatched,
			queryOptions.SortOrder,
		),
		userID,
		queryOptions.Offset,
		queryOptions.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// bind
	err = queries.Bind(rows, &watchlist)
	if err != nil {
		return nil, err
	}
	return watchlist, nil
}

func (repo *Repository) WatchlistCount(
	ctx context.Context,
	userID int,
	timeWatchedWhereClause string,
) (count int, err error) {
	// query
	row := repo.exec.QueryRowContext(
		ctx,
		fmt.Sprintf(watchfilmCountQuery, timeWatchedWhereClause),
		userID,
	)
	if err != nil {
		return 0, err
	}
	// bind
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *Repository) WatchlistAdd(
	ctx context.Context,
	userID int,
	filmID int,
) (watchID int, err error) {
	w := &models.Watchfilm{UserID: userID, FilmID: filmID}
	err = w.Insert(ctx, repo.exec, boil.Infer())
	if err != nil {
		return 0, err
	}
	return w.ID, nil
}

// As userID is not provided by the user, they cannot maliciously/inadvertently
// delete another user watchlist
func (repo *Repository) WatchlistDelete(
	ctx context.Context,
	userID int,
	watchID int,
) error {
	rowsAff, err := models.Watchfilms(
		models.WatchfilmWhere.ID.EQ(watchID),
		models.WatchfilmWhere.UserID.EQ(userID),
	).DeleteAll(ctx, repo.exec)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

// As userID is not provided by the user, they cannot maliciously/inadvertently
// delete another user watchlist
func (repo *Repository) WatchlistSetWatched(
	ctx context.Context,
	userID int,
	watchID int,
) error {
	rowsAff, err := models.Watchfilms(
		models.WatchfilmWhere.ID.EQ(watchID),
		models.WatchfilmWhere.UserID.EQ(userID),
	).UpdateAll(
		ctx, repo.exec, map[string]any{
			models.WatchfilmColumns.TimeWatched: null.TimeFrom(time.Now()),
		},
	)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}
