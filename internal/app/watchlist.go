package app

import (
	"context"

	"github.com/aria3ppp/watchlist-server/internal/query"
	"github.com/aria3ppp/watchlist-server/internal/repo"
	"github.com/aria3ppp/watchlist-server/internal/watchlist"
)

func (a *Application) WatchlistGet(
	ctx context.Context,
	userID int,
	queryOptions query.WatchlistOptions,
) (watchlist []*watchlist.Item, total int, err error) {
	// Run in a transaction context
	err = a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			var err error
			watchlist, err = tx.WatchlistGet(
				ctx,
				userID,
				queryOptions,
			)
			if err != nil {
				return err
			}
			total, err = tx.WatchlistCount(
				ctx,
				userID,
				queryOptions.WhereTimeWatched,
			)
			return err
		},
	)
	// This check is mandatory as if the transaction failed,
	// the saved results shouldn't return back to the caller
	if err != nil {
		return nil, 0, err
	}
	return watchlist, total, nil
}

func (a *Application) WatchlistAdd(
	ctx context.Context,
	userID int,
	filmID int,
) (watchID int, err error) {
	err = a.repository.Transaction(
		ctx,
		func(ctx context.Context, tx repo.Service) error {
			err := tx.FilmExists(ctx, filmID)
			if err != nil {
				if err == repo.ErrNoRecord {
					return ErrNotFound
				}
				return err
			}
			watchID, err = tx.WatchlistAdd(ctx, userID, filmID)
			return err
		},
	)
	if err != nil {
		return 0, err
	}
	return watchID, nil
}

func (a *Application) WatchlistDelete(
	ctx context.Context,
	userID int,
	watchID int,
) error {
	err := a.repository.WatchlistDelete(ctx, userID, watchID)
	if err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (a *Application) WatchlistSetWatched(
	ctx context.Context,
	userID int,
	watchID int,
) error {
	err := a.repository.WatchlistSetWatched(ctx, userID, watchID)
	if err != nil {
		if err == repo.ErrNoRecord {
			return ErrNotFound
		}
		return err
	}
	return nil
}
