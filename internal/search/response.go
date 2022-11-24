package search

import (
	"encoding/json"

	"github.com/aria3ppp/watchlist-server/internal/models"
)

type seriesHit models.Series

func (sh *seriesHit) UnmarshalJSON(data []byte) error {
	type Hit struct {
		Source models.Series `json:"_source"`
	}

	var hit Hit

	err := json.Unmarshal(data, &hit)
	if err != nil {
		return err
	}

	*sh = seriesHit(hit.Source)

	return nil
}

type movieHit models.Film

func (mh *movieHit) UnmarshalJSON(data []byte) error {
	type Hit struct {
		Source models.Film `json:"_source"`
	}

	var hit Hit

	err := json.Unmarshal(data, &hit)
	if err != nil {
		return err
	}

	*mh = movieHit(hit.Source)

	return nil
}
