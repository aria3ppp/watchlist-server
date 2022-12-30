package validator_test

import (
	"testing"

	"github.com/aria3ppp/watchlist-server/internal/validator"
	"github.com/stretchr/testify/require"
)

func TestIsPassword(t *testing.T) {
	testCases := []struct {
		name                           string
		nums, lowers, uppers, specials int
		passwd                         string
		err                            bool
	}{
		{"tc1", 2, 2, 2, 2, "54abcdEF&$", false},
		{"tc2", 2, 2, 2, 2, "", true},
		{"tc3", 0, 0, 0, 0, "fjdslj", false},
		{"tc4", 2, 2, 2, 2, "4abcdEF&$", true},
		{"tc5", 2, 2, 2, 2, "54aEF&$", true},
		{"tc6", 2, 2, 2, 2, "54abcdE&$", true},
		{"tc7", 2, 2, 2, 2, "54abcdEF&", true},
		{"tc8", 2, 0, 0, 0, "54aB$", false},
		{"tc9", 0, 2, 0, 0, "5abcdef", false},
		{"tc10", 0, 0, 2, 0, "5abcdEF", false},
		{"tc11", 0, 0, 0, 2, "5abcdef&$", false},
		{"tc12", 0, 0, 0, 0, "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			r := validator.IsPassword().
				Numbers(tc.nums).
				LowerLetters(tc.lowers).
				UpperLetters(tc.uppers).
				SpecialChars(tc.specials)

			err := r.Validate(tc.passwd)
			if tc.err {
				expValidationError := validator.ErrInvalidPassword.SetParams(
					map[string]any{
						"num":     tc.nums,
						"lower":   tc.lowers,
						"upper":   tc.uppers,
						"special": tc.specials,
					},
				)
				require.Equal(expValidationError, err)
			} else {
				require.NoError(err)
			}
		})
	}
}
