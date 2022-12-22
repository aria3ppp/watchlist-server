package dto_test

import (
	"log"
	"path/filepath"

	"github.com/aria3ppp/watchlist-server/internal/config"
)

func init() {
	if err := config.Load(filepath.Join("..", "..", "config.yml")); err != nil {
		log.Fatalf("dto.init: faild loading config file: %s", err)
	}
}
