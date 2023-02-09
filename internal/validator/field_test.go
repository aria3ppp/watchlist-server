package validator_test

import (
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/validator"
	"github.com/stretchr/testify/require"
)

func TestIsValidFieldOfModel(t *testing.T) {
	testCases := []struct {
		name  string
		model string
		field string
		err   bool
	}{
		{"tc1", models.TableNames.Films, "non-existent-field", true},
		{"tc2", models.TableNames.Films, models.SeriesColumns.DateEnded, true},
		{
			"tc3",
			models.TableNames.Serieses,
			models.SeriesColumns.DateEnded,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			r := validator.IsValidFieldOfModel(tc.model)

			err := r.Validate(tc.field)
			if tc.err {
				expValidationError := validator.ErrInvalidFieldOfModel.SetParams(
					map[string]any{
						"model": tc.model,
						"field": tc.field,
					},
				)
				require.Equal(expValidationError, err)
			} else {
				require.NoError(err)
			}
		})
	}
}
