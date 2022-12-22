package watchlist

import "github.com/aria3ppp/watchlist-server/internal/models"

// sync `boil` tag with model name everytime there's a change
type Item struct {
	models.Watchfilm `boil:"watchfilms,bind"`
	Film             models.Film `boil:"films,bind"      json:"film"`
}
