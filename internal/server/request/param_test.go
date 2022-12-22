package request_test

import (
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/server/request"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/require"
)

func TestSeriesSeasonEpisodeNumberPathParam_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		params   request.SeriesSeasonEpisodeNumberPathParam
		expError error
	}{
		{
			name:   "tc1",
			params: request.SeriesSeasonEpisodeNumberPathParam{},
			expError: validation.Errors{
				"id":             validation.ErrRequired,
				"season_number":  validation.ErrRequired,
				"episode_number": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			params: request.SeriesSeasonEpisodeNumberPathParam{
				SeriesID:      -1,
				SeasonNumber:  -1,
				EpisodeNumber: -1,
			},
			expError: validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"episode_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			},
		},
		{
			name: "tc3",
			params: request.SeriesSeasonEpisodeNumberPathParam{
				SeriesID:      1,
				SeasonNumber:  config.Config.Validation.Film.SeasonNumber.MaxValue + 1,
				EpisodeNumber: config.Config.Validation.Film.EpisodeNumber.MaxValue + 1,
			},
			expError: validation.Errors{
				"season_number": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Film.SeasonNumber.MaxValue,
					},
				),
				"episode_number": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Film.EpisodeNumber.MaxValue,
					},
				),
			},
		},
		{
			name: "tc4",
			params: request.SeriesSeasonEpisodeNumberPathParam{
				SeriesID:      1,
				SeasonNumber:  1,
				EpisodeNumber: 1,
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

func TestSeriesSeasonNumberPathParam_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		params   request.SeriesSeasonNumberPathParam
		expError error
	}{
		{
			name:   "tc1",
			params: request.SeriesSeasonNumberPathParam{},
			expError: validation.Errors{
				"id":            validation.ErrRequired,
				"season_number": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			params: request.SeriesSeasonNumberPathParam{
				SeriesID:     -1,
				SeasonNumber: -1,
			},
			expError: validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
				"season_number": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			},
		},
		{
			name: "tc3",
			params: request.SeriesSeasonNumberPathParam{
				SeriesID:     1,
				SeasonNumber: config.Config.Validation.Film.SeasonNumber.MaxValue + 1,
			},
			expError: validation.Errors{
				"season_number": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Film.SeasonNumber.MaxValue,
					},
				),
			},
		},
		{
			name: "tc4",
			params: request.SeriesSeasonNumberPathParam{
				SeriesID:     1,
				SeasonNumber: 1,
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

func TestIDPathParam_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		params   request.IDPathParam
		expError error
	}{
		{
			name:   "tc1",
			params: request.IDPathParam{},
			expError: validation.Errors{
				"id": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			params: request.IDPathParam{
				ID: -1,
			},
			expError: validation.Errors{
				"id": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{"threshold": 1},
				),
			},
		},
		{
			name: "tc3",
			params: request.IDPathParam{
				ID: 1,
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
