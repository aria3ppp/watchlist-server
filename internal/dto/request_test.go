package dto_test

import (
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/dto"
	"github.com/aria3ppp/watchlist-server/internal/testutils"
	"github.com/aria3ppp/watchlist-server/internal/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func TestUserCreateRequest_Validate(t *testing.T) {
	timeNow := time.Now()
	testCases := []struct {
		name     string
		req      dto.UserCreateRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.UserCreateRequest{},
			expError: validation.Errors{
				"email":    validation.ErrRequired,
				"password": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.UserCreateRequest{
				Email:    "email@example.com",
				Password: "pa$$W0RD0",
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.UserCreateRequest{
				Email:    "invalid_email",
				Password: "invalid_password",
			},
			expError: validation.Errors{
				"email": is.ErrEmail,
				"password": validator.ErrInvalidPassword.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
			},
		},
		{
			name: "tc4",
			req: dto.UserCreateRequest{
				Email:     "e",
				Password:  "p",
				FirstName: null.StringFrom("f"),
				LastName:  null.StringFrom("l"),
				Bio:       null.StringFrom("b"),
				Birthdate: null.TimeFrom(testutils.Date(
					config.Config.Validation.User.Birthdate.MinValue.Year-1,
					time.Month(
						config.Config.Validation.User.Birthdate.MinValue.Month,
					),
					config.Config.Validation.User.Birthdate.MinValue.Day,
				)),
			},
			expError: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
				"last_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.LastName.MinLength,
						"max": config.Config.Validation.User.LastName.MaxLength,
					},
				),
				"bio": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Bio.MinLength,
						"max": config.Config.Validation.User.Bio.MaxLength,
					},
				),
				"birthdate": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.User.Birthdate.MinValue.Year,
							time.Month(
								config.Config.Validation.User.Birthdate.MinValue.Month,
							),
							config.Config.Validation.User.Birthdate.MinValue.Day,
						),
					},
				),
			},
		},
		{
			name: "tc5",
			req: dto.UserCreateRequest{
				Email: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Email.MaxLength,
				),
				Password: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
				FirstName: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.User.FirstName.MaxLength,
					),
				),
				LastName: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.User.LastName.MaxLength,
					),
				),
				Bio: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.User.Bio.MaxLength,
					),
				),
				Birthdate: null.TimeFrom(testutils.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
				).Add(time.Hour)),
			},
			expError: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
				"last_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.LastName.MinLength,
						"max": config.Config.Validation.User.LastName.MaxLength,
					},
				),
				"bio": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Bio.MinLength,
						"max": config.Config.Validation.User.Bio.MaxLength,
					},
				),
				"birthdate": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							timeNow.Year(),
							timeNow.Month(),
							timeNow.Day(),
						),
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestUserLoginRequest_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		req      dto.UserLoginRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.UserLoginRequest{},
			expError: validation.Errors{
				"email":    validation.ErrRequired,
				"password": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.UserLoginRequest{
				Email:    "email@example.com",
				Password: "pa$$W0RD0",
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.UserLoginRequest{
				Email:    "invalid_email",
				Password: "invalid_password",
			},
			expError: validation.Errors{
				"email": is.ErrEmail,
				"password": validator.ErrInvalidPassword.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
			},
		},
		{
			name: "tc4",
			req: dto.UserLoginRequest{
				Email:    "e",
				Password: "p",
			},
			expError: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},
		{
			name: "tc5",
			req: dto.UserLoginRequest{
				Email: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Email.MaxLength,
				),
				Password: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
			},
			expError: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestUserUpdateRequest_Validate(t *testing.T) {
	timeNow := time.Now()
	testCases := []struct {
		name     string
		req      dto.UserUpdateRequest
		expError error
	}{
		{
			name:     "tc1",
			req:      dto.UserUpdateRequest{},
			expError: nil,
		},
		{
			name: "tc2",
			req: dto.UserUpdateRequest{
				FirstName: null.StringFrom("f"),
				LastName:  null.StringFrom("l"),
				Bio:       null.StringFrom("b"),
				Birthdate: null.TimeFrom(testutils.Date(
					config.Config.Validation.User.Birthdate.MinValue.Year-1,
					time.Month(
						config.Config.Validation.User.Birthdate.MinValue.Month,
					),
					config.Config.Validation.User.Birthdate.MinValue.Day,
				)),
			},
			expError: validation.Errors{
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
				"last_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.LastName.MinLength,
						"max": config.Config.Validation.User.LastName.MaxLength,
					},
				),
				"bio": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Bio.MinLength,
						"max": config.Config.Validation.User.Bio.MaxLength,
					},
				),
				"birthdate": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.User.Birthdate.MinValue.Year,
							time.Month(
								config.Config.Validation.User.Birthdate.MinValue.Month,
							),
							config.Config.Validation.User.Birthdate.MinValue.Day,
						),
					},
				),
			},
		},
		{
			name: "tc3",
			req: dto.UserUpdateRequest{
				FirstName: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.User.FirstName.MaxLength,
					),
				),
				LastName: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.User.LastName.MaxLength,
					),
				),
				Bio: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.User.Bio.MaxLength,
					),
				),
				Birthdate: null.TimeFrom(testutils.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
				).Add(time.Hour)),
			},
			expError: validation.Errors{
				"first_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.FirstName.MinLength,
						"max": config.Config.Validation.User.FirstName.MaxLength,
					},
				),
				"last_name": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.LastName.MinLength,
						"max": config.Config.Validation.User.LastName.MaxLength,
					},
				),
				"bio": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Bio.MinLength,
						"max": config.Config.Validation.User.Bio.MaxLength,
					},
				),
				"birthdate": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							timeNow.Year(),
							timeNow.Month(),
							timeNow.Day(),
						),
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestUserEmailUpdateRequest_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		req      dto.UserEmailUpdateRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.UserEmailUpdateRequest{},
			expError: validation.Errors{
				"email": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.UserEmailUpdateRequest{
				Email: "email@example.com",
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.UserEmailUpdateRequest{
				Email: "invalid_email",
			},
			expError: validation.Errors{
				"email": is.ErrEmail,
			},
		},
		{
			name: "tc4",
			req: dto.UserEmailUpdateRequest{
				Email: "e",
			},
			expError: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
			},
		},
		{
			name: "tc5",
			req: dto.UserEmailUpdateRequest{
				Email: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Email.MaxLength,
				),
			},
			expError: validation.Errors{
				"email": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Email.MinLength,
						"max": config.Config.Validation.User.Email.MaxLength,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestUserPasswordUpdateRequest_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		req      dto.UserPasswordUpdateRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.UserPasswordUpdateRequest{},
			expError: validation.Errors{
				"current_password": validation.ErrRequired,
				"new_password":     validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.UserPasswordUpdateRequest{
				CurrentPassword: "pa$$W0RD0",
				NewPassword:     "new_pa$$W0RD0",
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.UserPasswordUpdateRequest{
				CurrentPassword: "invalid_password",
				NewPassword:     "new_invalid_password",
			},
			expError: validation.Errors{
				"current_password": validator.ErrInvalidPassword.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
				"new_password": validator.ErrInvalidPassword.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
			},
		},
		{
			name: "tc4",
			req: dto.UserPasswordUpdateRequest{
				CurrentPassword: "c",
				NewPassword:     "n",
			},
			expError: validation.Errors{
				"current_password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
				"new_password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},
		{
			name: "tc5",
			req: dto.UserPasswordUpdateRequest{
				CurrentPassword: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
				NewPassword: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
			},
			expError: validation.Errors{
				"current_password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
				"new_password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestUserDeleteRequest_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		req      dto.UserDeleteRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.UserDeleteRequest{},
			expError: validation.Errors{
				"password": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.UserDeleteRequest{
				Password: "pa$$W0RD0",
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.UserDeleteRequest{
				Password: "invalid_password",
			},
			expError: validation.Errors{
				"password": validator.ErrInvalidPassword.SetParams(
					map[string]any{
						"num":     config.Config.Validation.User.Password.RequiredNumbers,
						"lower":   config.Config.Validation.User.Password.RequiredLowerLetters,
						"upper":   config.Config.Validation.User.Password.RequiredUpperLetters,
						"special": config.Config.Validation.User.Password.RequiredSpecialChars,
					},
				),
			},
		},
		{
			name: "tc4",
			req: dto.UserDeleteRequest{
				Password: "p",
			},
			expError: validation.Errors{
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},
		{
			name: "tc5",
			req: dto.UserDeleteRequest{
				Password: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.User.Password.MaxLength,
				),
			},
			expError: validation.Errors{
				"password": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.User.Password.MinLength,
						"max": config.Config.Validation.User.Password.MaxLength,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestSeriesCreateRequest_Validate(t *testing.T) {
	timeNow := time.Now()
	testCases := []struct {
		name     string
		req      dto.SeriesCreateRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.SeriesCreateRequest{},
			expError: validation.Errors{
				"title":        validation.ErrRequired,
				"date_started": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.SeriesCreateRequest{
				Title:       "Breaking Bad",
				DateStarted: testutils.Date(2008, time.January, 20),
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.SeriesCreateRequest{
				Title:        "t",
				Descriptions: null.StringFrom("d"),
				DateStarted: testutils.Date(
					config.Config.Validation.Series.DateStarted.MinValue.Year-1,
					time.Month(
						config.Config.Validation.Series.DateStarted.MinValue.Month,
					),
					config.Config.Validation.Series.DateStarted.MinValue.Day,
				),
				DateEnded: null.TimeFrom(testutils.Date(
					config.Config.Validation.Series.DateEnded.MinValue.Year-1,
					time.Month(
						config.Config.Validation.Series.DateEnded.MinValue.Month,
					),
					config.Config.Validation.Series.DateEnded.MinValue.Day,
				)),
			},
			expError: validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Title.MinLength,
						"max": config.Config.Validation.Series.Title.MaxLength,
					},
				),
				"descriptions": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Descriptions.MinLength,
						"max": config.Config.Validation.Series.Descriptions.MaxLength,
					},
				),
				"date_started": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.Series.DateStarted.MinValue.Year,
							time.Month(
								config.Config.Validation.Series.DateStarted.MinValue.Month,
							),
							config.Config.Validation.Series.DateStarted.MinValue.Day,
						),
					},
				),
				"date_ended": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.Series.DateEnded.MinValue.Year,
							time.Month(
								config.Config.Validation.Series.DateEnded.MinValue.Month,
							),
							config.Config.Validation.Series.DateEnded.MinValue.Day,
						),
					},
				),
			},
		},
		{
			name: "tc4",
			req: dto.SeriesCreateRequest{
				Title: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.Series.Title.MaxLength,
				),
				Descriptions: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.Series.Descriptions.MaxLength,
					),
				),
				DateStarted: testutils.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
				).Add(time.Hour),
				DateEnded: null.TimeFrom(testutils.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
				).Add(time.Hour)),
			},
			expError: validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Title.MinLength,
						"max": config.Config.Validation.Series.Title.MaxLength,
					},
				),
				"descriptions": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Descriptions.MinLength,
						"max": config.Config.Validation.Series.Descriptions.MaxLength,
					},
				),
				"date_started": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							timeNow.Year(),
							timeNow.Month(),
							timeNow.Day(),
						),
					},
				),
				"date_ended": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							timeNow.Year(),
							timeNow.Month(),
							timeNow.Day(),
						),
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestSeriesUpdateRequest_Validate(t *testing.T) {
	timeNow := time.Now()
	testCases := []struct {
		name     string
		req      dto.SeriesUpdateRequest
		expError error
	}{
		{
			name:     "tc1",
			req:      dto.SeriesUpdateRequest{},
			expError: nil,
		},
		{
			name: "tc2",
			req: dto.SeriesUpdateRequest{
				Title:        null.StringFrom("t"),
				Descriptions: null.StringFrom("d"),
				DateStarted: null.TimeFrom(testutils.Date(
					config.Config.Validation.Series.DateStarted.MinValue.Year-1,
					time.Month(
						config.Config.Validation.Series.DateStarted.MinValue.Month,
					),
					config.Config.Validation.Series.DateStarted.MinValue.Day,
				)),
				DateEnded: null.TimeFrom(testutils.Date(
					config.Config.Validation.Series.DateEnded.MinValue.Year-1,
					time.Month(
						config.Config.Validation.Series.DateEnded.MinValue.Month,
					),
					config.Config.Validation.Series.DateEnded.MinValue.Day,
				)),
			},
			expError: validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Title.MinLength,
						"max": config.Config.Validation.Series.Title.MaxLength,
					},
				),
				"descriptions": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Descriptions.MinLength,
						"max": config.Config.Validation.Series.Descriptions.MaxLength,
					},
				),
				"date_started": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.Series.DateStarted.MinValue.Year,
							time.Month(
								config.Config.Validation.Series.DateStarted.MinValue.Month,
							),
							config.Config.Validation.Series.DateStarted.MinValue.Day,
						),
					},
				),
				"date_ended": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.Series.DateEnded.MinValue.Year,
							time.Month(
								config.Config.Validation.Series.DateEnded.MinValue.Month,
							),
							config.Config.Validation.Series.DateEnded.MinValue.Day,
						),
					},
				),
			},
		},
		{
			name: "tc3",
			req: dto.SeriesUpdateRequest{
				Title: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.Series.Title.MaxLength,
					),
				),
				Descriptions: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.Series.Descriptions.MaxLength,
					),
				),
				DateStarted: null.TimeFrom(testutils.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
				).Add(time.Hour)),
				DateEnded: null.TimeFrom(testutils.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
				).Add(time.Hour)),
			},
			expError: validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Title.MinLength,
						"max": config.Config.Validation.Series.Title.MaxLength,
					},
				),
				"descriptions": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Series.Descriptions.MinLength,
						"max": config.Config.Validation.Series.Descriptions.MaxLength,
					},
				),
				"date_started": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							timeNow.Year(),
							timeNow.Month(),
							timeNow.Day(),
						),
					},
				),
				"date_ended": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							timeNow.Year(),
							timeNow.Month(),
							timeNow.Day(),
						),
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestFilmCreateRequest_Validate(t *testing.T) {
	timeNow := time.Now()
	testCases := []struct {
		name     string
		req      dto.FilmCreateRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.FilmCreateRequest{},
			expError: validation.Errors{
				"title":         validation.ErrRequired,
				"date_released": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.FilmCreateRequest{
				Title:        "Film",
				DateReleased: testutils.Date(2000, time.November, 11),
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.FilmCreateRequest{
				Title:        "t",
				Descriptions: null.StringFrom("d"),
				DateReleased: testutils.Date(
					config.Config.Validation.Film.DateReleased.MinValue.Year-1,
					time.Month(
						config.Config.Validation.Film.DateReleased.MinValue.Month,
					),
					config.Config.Validation.Film.DateReleased.MinValue.Day,
				),
				Duration: null.IntFrom(
					config.Config.Validation.Film.Duraion.MinLength - 1,
				),
			},
			expError: validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Title.MinLength,
						"max": config.Config.Validation.Film.Title.MaxLength,
					},
				),
				"descriptions": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Descriptions.MinLength,
						"max": config.Config.Validation.Film.Descriptions.MaxLength,
					},
				),
				"date_released": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.Film.DateReleased.MinValue.Year,
							time.Month(
								config.Config.Validation.Film.DateReleased.MinValue.Month,
							),
							config.Config.Validation.Film.DateReleased.MinValue.Day,
						),
					},
				),
				"duration": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Film.Duraion.MinLength,
					},
				),
			},
		},
		{
			name: "tc4",
			req: dto.FilmCreateRequest{
				Title: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.Film.Title.MaxLength,
				),
				Descriptions: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.Film.Descriptions.MaxLength,
					),
				),
				DateReleased: testutils.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
				).Add(time.Hour),
				Duration: null.IntFrom(
					config.Config.Validation.Film.Duraion.MaxLength + 1,
				),
			},
			expError: validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Title.MinLength,
						"max": config.Config.Validation.Film.Title.MaxLength,
					},
				),
				"descriptions": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Descriptions.MinLength,
						"max": config.Config.Validation.Film.Descriptions.MaxLength,
					},
				),
				"date_released": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							timeNow.Year(),
							timeNow.Month(),
							timeNow.Day(),
						),
					},
				),
				"duration": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Film.Duraion.MaxLength,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestFilmUpdateRequest_Validate(t *testing.T) {
	timeNow := time.Now()
	testCases := []struct {
		name     string
		req      dto.FilmUpdateRequest
		expError error
	}{
		{
			name:     "tc1",
			req:      dto.FilmUpdateRequest{},
			expError: nil,
		},
		{
			name: "tc2",
			req: dto.FilmUpdateRequest{
				Title:        null.StringFrom("t"),
				Descriptions: null.StringFrom("d"),
				DateReleased: null.TimeFrom(testutils.Date(
					config.Config.Validation.Film.DateReleased.MinValue.Year-1,
					time.Month(
						config.Config.Validation.Film.DateReleased.MinValue.Month,
					),
					config.Config.Validation.Film.DateReleased.MinValue.Day,
				)),
				Duration: null.IntFrom(
					config.Config.Validation.Film.Duraion.MinLength - 1,
				),
			},
			expError: validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Title.MinLength,
						"max": config.Config.Validation.Film.Title.MaxLength,
					},
				),
				"descriptions": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Descriptions.MinLength,
						"max": config.Config.Validation.Film.Descriptions.MaxLength,
					},
				),
				"date_released": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							config.Config.Validation.Film.DateReleased.MinValue.Year,
							time.Month(
								config.Config.Validation.Film.DateReleased.MinValue.Month,
							),
							config.Config.Validation.Film.DateReleased.MinValue.Day,
						),
					},
				),
				"duration": validation.ErrMinGreaterEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Film.Duraion.MinLength,
					},
				),
			},
		},
		{
			name: "tc3",
			req: dto.FilmUpdateRequest{
				Title: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.Film.Title.MaxLength,
					),
				),
				Descriptions: null.StringFrom(
					testutils.GenerateStringLongerThanMaxLength(
						config.Config.Validation.Film.Descriptions.MaxLength,
					),
				),
				DateReleased: null.TimeFrom(testutils.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
				).Add(time.Hour)),
				Duration: null.IntFrom(
					config.Config.Validation.Film.Duraion.MaxLength + 1,
				),
			},
			expError: validation.Errors{
				"title": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Title.MinLength,
						"max": config.Config.Validation.Film.Title.MaxLength,
					},
				),
				"descriptions": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Film.Descriptions.MinLength,
						"max": config.Config.Validation.Film.Descriptions.MaxLength,
					},
				),
				"date_released": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": testutils.Date(
							timeNow.Year(),
							timeNow.Month(),
							timeNow.Day(),
						),
					},
				),
				"duration": validation.ErrMaxLessEqualThanRequired.SetParams(
					map[string]any{
						"threshold": config.Config.Validation.Film.Duraion.MaxLength,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestEpisodesPutAllBySeasonRequest_Validate(t *testing.T) {
	timeNow := time.Now()
	testCases := []struct {
		name     string
		req      dto.EpisodesPutAllBySeasonRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.EpisodesPutAllBySeasonRequest{},
			expError: validation.Errors{
				"episodes": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.EpisodesPutAllBySeasonRequest{
				Episodes: []*dto.EpisodePutRequest{
					{
						Title:        "Episode",
						DateReleased: testutils.Date(2000, time.November, 11),
					},
				},
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.EpisodesPutAllBySeasonRequest{
				Episodes: []*dto.EpisodePutRequest{
					{},
					{
						Title:        "Episode",
						DateReleased: testutils.Date(2000, time.November, 11),
					},
					{
						Title:        "t",
						Descriptions: null.StringFrom("d"),
						DateReleased: testutils.Date(
							config.Config.Validation.Film.DateReleased.MinValue.Year-1,
							time.Month(
								config.Config.Validation.Film.DateReleased.MinValue.Month,
							),
							config.Config.Validation.Film.DateReleased.MinValue.Day,
						),
						Duration: null.IntFrom(
							config.Config.Validation.Film.Duraion.MinLength - 1,
						),
					},
					{
						Title: testutils.GenerateStringLongerThanMaxLength(
							config.Config.Validation.Film.Title.MaxLength,
						),
						Descriptions: null.StringFrom(
							testutils.GenerateStringLongerThanMaxLength(
								config.Config.Validation.Film.Descriptions.MaxLength,
							),
						),
						DateReleased: testutils.Date(
							timeNow.Year(), timeNow.Month(), timeNow.Day(),
						).Add(time.Hour),
						Duration: null.IntFrom(
							config.Config.Validation.Film.Duraion.MaxLength + 1,
						),
					},
				},
			},
			expError: validation.Errors{
				"episodes": validation.Errors{
					"0": validation.Errors{
						"title":         validation.ErrRequired,
						"date_released": validation.ErrRequired,
					},
					"2": validation.Errors{
						"title": validation.ErrLengthOutOfRange.SetParams(
							map[string]any{
								"min": config.Config.Validation.Film.Title.MinLength,
								"max": config.Config.Validation.Film.Title.MaxLength,
							},
						),
						"descriptions": validation.ErrLengthOutOfRange.SetParams(
							map[string]any{
								"min": config.Config.Validation.Film.Descriptions.MinLength,
								"max": config.Config.Validation.Film.Descriptions.MaxLength,
							},
						),
						"date_released": validation.ErrMinGreaterEqualThanRequired.SetParams(
							map[string]any{
								"threshold": testutils.Date(
									config.Config.Validation.Film.DateReleased.MinValue.Year,
									time.Month(
										config.Config.Validation.Film.DateReleased.MinValue.Month,
									),
									config.Config.Validation.Film.DateReleased.MinValue.Day,
								),
							},
						),
						"duration": validation.ErrMinGreaterEqualThanRequired.SetParams(
							map[string]any{
								"threshold": config.Config.Validation.Film.Duraion.MinLength,
							},
						),
					},
					"3": validation.Errors{
						"title": validation.ErrLengthOutOfRange.SetParams(
							map[string]any{
								"min": config.Config.Validation.Film.Title.MinLength,
								"max": config.Config.Validation.Film.Title.MaxLength,
							},
						),
						"descriptions": validation.ErrLengthOutOfRange.SetParams(
							map[string]any{
								"min": config.Config.Validation.Film.Descriptions.MinLength,
								"max": config.Config.Validation.Film.Descriptions.MaxLength,
							},
						),
						"date_released": validation.ErrMaxLessEqualThanRequired.SetParams(
							map[string]any{
								"threshold": testutils.Date(
									timeNow.Year(),
									timeNow.Month(),
									timeNow.Day(),
								),
							},
						),
						"duration": validation.ErrMaxLessEqualThanRequired.SetParams(
							map[string]any{
								"threshold": config.Config.Validation.Film.Duraion.MaxLength,
							},
						),
					},
				},
			},
		},
		{
			name: "tc4",
			req: dto.EpisodesPutAllBySeasonRequest{
				Episodes: func() (episodes []*dto.EpisodePutRequest) {
					for i := 0; i < config.Config.Validation.Request.Array.MaxLength+1; i++ {
						episodes = append(episodes, &dto.EpisodePutRequest{})
					}
					return
				}(),
			},
			expError: validation.Errors{
				"episodes": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": 1,
						"max": config.Config.Validation.Request.Array.MaxLength,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}

func TestInvalidationRequest_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		req      dto.InvalidationRequest
		expError error
	}{
		{
			name: "tc1",
			req:  dto.InvalidationRequest{},
			expError: validation.Errors{
				"invalidation": validation.ErrRequired,
			},
		},
		{
			name: "tc2",
			req: dto.InvalidationRequest{
				Invalidation: "Invalidation note!",
			},
			expError: nil,
		},
		{
			name: "tc3",
			req: dto.InvalidationRequest{
				Invalidation: "i",
			},
			expError: validation.Errors{
				"invalidation": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Request.Invalidation.MinLength,
						"max": config.Config.Validation.Request.Invalidation.MaxLength,
					},
				),
			},
		},
		{
			name: "tc4",
			req: dto.InvalidationRequest{
				Invalidation: testutils.GenerateStringLongerThanMaxLength(
					config.Config.Validation.Request.Invalidation.MaxLength,
				),
			},
			expError: validation.Errors{
				"invalidation": validation.ErrLengthOutOfRange.SetParams(
					map[string]any{
						"min": config.Config.Validation.Request.Invalidation.MinLength,
						"max": config.Config.Validation.Request.Invalidation.MaxLength,
					},
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			require.Equal(tc.expError, tc.req.Validate())
		})
	}
}
