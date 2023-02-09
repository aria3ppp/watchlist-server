package watchlist

import "github.com/aria3ppp/watchlist-server/internal/models"

// sync `boil` tag whenever there's a change in model name
type Item struct {
	models.Watchfilm `boil:"watchfilms,bind"`
	Film             models.Film `boil:"films,bind"      json:"film"`
}
