package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	testCases := []struct {
		name                           string
		nums, lowers, uppers, specials int
		passwd                         string
		err                            bool
	}{
		{"tc1", 2, 2, 2, 2, "54abcdEF&$", false},
		{"tc2", 2, 2, 2, 2, "", false},
		{"tc8", 0, 0, 0, 0, "fjdslj", false},
		{"tc3", 2, 2, 2, 2, "4abcdEF&$", true},
		{"tc4", 2, 2, 2, 2, "54aEF&$", true},
		{"tc5", 2, 2, 2, 2, "54abcdE&$", true},
		{"tc6", 2, 2, 2, 2, "54abcdEF&", true},
		{"tc7", 2, 0, 0, 0, "54abcdef", false},
		{"tc7", 0, 2, 0, 0, "54abcdef", false},
		{"tc7", 0, 0, 2, 0, "54abcdEF", false},
		{"tc7", 0, 0, 0, 2, "54abcdef&$", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			r := Password(tc.nums, tc.lowers, tc.uppers, tc.specials)
			err := r.Validate(tc.passwd)
			if tc.err {
				expValidationError := ErrPasswordInvalid.SetParams(
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
