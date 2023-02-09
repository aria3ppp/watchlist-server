package request

import (
	"github.com/aria3ppp/watchlist-server/internal/config"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type IDPathParam struct {
	ID int `param:"id" json:"id"`
}

var _ validation.Validatable = IDPathParam{}

func (p IDPathParam) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(
			&p.ID,
			validation.Required,
			validation.Min(1),
		),
	)
}

type SeriesSeasonNumberPathParam struct {
	SeriesID     int `param:"id"            json:"id"`
	SeasonNumber int `param:"season_number" json:"season_number"`
}

var _ validation.Validatable = SeriesSeasonNumberPathParam{}

func (p SeriesSeasonNumberPathParam) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(
			&p.SeriesID,
			validation.Required,
			validation.Min(1),
		),
		validation.Field(
			&p.SeasonNumber,
			validation.Required,
			validation.Min(1),
			validation.Max(config.Config.Validation.Film.SeasonNumber.MaxValue),
		),
	)
}

type SeriesSeasonEpisodeNumberPathParam struct {
	SeriesID      int `param:"id"             json:"id"`
	SeasonNumber  int `param:"season_number"  json:"season_number"`
	EpisodeNumber int `param:"episode_number" json:"episode_number"`
}

var _ validation.Validatable = SeriesSeasonEpisodeNumberPathParam{}

func (p SeriesSeasonEpisodeNumberPathParam) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(
			&p.SeriesID,
			validation.Required,
			validation.Min(1),
		),
		validation.Field(
			&p.SeasonNumber,
			validation.Required,
			validation.Min(1),
			validation.Max(config.Config.Validation.Film.SeasonNumber.MaxValue),
		),
		validation.Field(
			&p.EpisodeNumber,
			validation.Required,
			validation.Min(1),
			validation.Max(
				config.Config.Validation.Film.EpisodeNumber.MaxValue,
			),
		),
	)
}
