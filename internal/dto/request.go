package dto

import (
	"time"

	"github.com/aria3ppp/watchlist-server/internal/config"
	"github.com/aria3ppp/watchlist-server/internal/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null/v8"
)

var (
	emailValidationRules = []validation.Rule{
		validation.Required,
		validation.Length(
			config.Config.Validation.User.Email.MinLength,
			config.Config.Validation.User.Email.MaxLength,
		),
		is.EmailFormat,
	}

	passwordValidationRules = []validation.Rule{
		validation.Required,
		validation.Length(
			config.Config.Validation.User.Password.MinLength,
			config.Config.Validation.User.Password.MaxLength,
		),
		validator.IsPassword().
			Numbers(config.Config.Validation.User.Password.RequiredNumbers).
			LowerLetters(config.Config.Validation.User.Password.RequiredLowerLetters).
			UpperLetters(config.Config.Validation.User.Password.RequiredUpperLetters).
			SpecialChars(config.Config.Validation.User.Password.RequiredSpecialChars),
	}
)

// -----------------------------------------------------------------------------
// UserCreateRequest
// -----------------------------------------------------------------------------
type UserCreateRequest struct {
	Email     string      `json:"email"`
	Password  string      `json:"password"`
	FirstName null.String `json:"first_name"`
	LastName  null.String `json:"last_name"`
	Bio       null.String `json:"bio"`
	Birthdate null.Time   `json:"birthdate"`
}

var _ validation.Validatable = UserCreateRequest{}

func (r UserCreateRequest) Validate() error {
	timeNow := time.Now()

	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Email,
			emailValidationRules...,
		),
		validation.Field(
			&r.Password,
			passwordValidationRules...,
		),
		validation.Field(
			&r.FirstName,
			validation.When(
				r.FirstName.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.User.FirstName.MinLength,
					config.Config.Validation.User.FirstName.MaxLength,
				),
			),
		),
		validation.Field(
			&r.LastName,
			validation.When(
				r.LastName.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.User.LastName.MinLength,
					config.Config.Validation.User.LastName.MaxLength,
				),
			),
		),
		validation.Field(
			&r.Bio,
			validation.When(
				r.Bio.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.User.Bio.MinLength,
					config.Config.Validation.User.Bio.MaxLength,
				),
			),
		),
		validation.Field(
			&r.Birthdate,
			validation.When(
				r.Birthdate.Valid,
				validation.Required,
				validation.Min(
					time.Date(
						config.Config.Validation.User.Birthdate.MinValue.Year,
						time.Month(
							config.Config.Validation.User.Birthdate.MinValue.Month,
						),
						config.Config.Validation.User.Birthdate.MinValue.Day,
						0, 0, 0, 0, time.UTC,
					),
				),
				validation.Max(
					time.Date(
						timeNow.Year(), timeNow.Month(), timeNow.Day(),
						0, 0, 0, 0, time.UTC,
					),
				),
			),
		),
	)
}

// -----------------------------------------------------------------------------
// UserLoginRequest
// -----------------------------------------------------------------------------
type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var _ validation.Validatable = UserLoginRequest{}

func (r UserLoginRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Email,
			emailValidationRules...,
		),
		validation.Field(
			&r.Password,
			passwordValidationRules...,
		),
	)
}

// -----------------------------------------------------------------------------
// UserUpdateRequest
// -----------------------------------------------------------------------------
type UserUpdateRequest struct {
	FirstName null.String `json:"first_name"`
	LastName  null.String `json:"last_name"`
	Bio       null.String `json:"bio"`
	Birthdate null.Time   `json:"birthdate"`
}

var _ validation.Validatable = UserUpdateRequest{}

func (r UserUpdateRequest) Validate() error {
	timeNow := time.Now()

	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.FirstName,
			validation.When(
				r.FirstName.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.User.FirstName.MinLength,
					config.Config.Validation.User.FirstName.MaxLength,
				),
			),
		),
		validation.Field(
			&r.LastName,
			validation.When(
				r.LastName.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.User.LastName.MinLength,
					config.Config.Validation.User.LastName.MaxLength,
				),
			),
		),
		validation.Field(
			&r.Bio,
			validation.When(
				r.Bio.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.User.Bio.MinLength,
					config.Config.Validation.User.Bio.MaxLength,
				),
			),
		),
		validation.Field(
			&r.Birthdate,
			validation.When(
				r.Birthdate.Valid,
				validation.Required,
				validation.Min(
					time.Date(
						config.Config.Validation.User.Birthdate.MinValue.Year,
						time.Month(
							config.Config.Validation.User.Birthdate.MinValue.Month,
						),
						config.Config.Validation.User.Birthdate.MinValue.Day,
						0, 0, 0, 0, time.UTC,
					),
				),
				validation.Max(
					time.Date(
						timeNow.Year(), timeNow.Month(), timeNow.Day(),
						0, 0, 0, 0, time.UTC,
					),
				),
			),
		),
	)
}

// -----------------------------------------------------------------------------
// UserEmailUpdateRequest
// -----------------------------------------------------------------------------
type UserEmailUpdateRequest struct {
	Email string `json:"email"`
}

var _ validation.Validatable = UserEmailUpdateRequest{}

func (r UserEmailUpdateRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Email,
			emailValidationRules...,
		),
	)
}

// -----------------------------------------------------------------------------
// UserPasswordUpdateRequest
// -----------------------------------------------------------------------------
type UserPasswordUpdateRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

var _ validation.Validatable = UserPasswordUpdateRequest{}

func (r UserPasswordUpdateRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.CurrentPassword,
			passwordValidationRules...,
		),
		validation.Field(
			&r.NewPassword,
			passwordValidationRules...,
		),
	)
}

// -----------------------------------------------------------------------------
// UserDeleteRequest
// -----------------------------------------------------------------------------
type UserDeleteRequest struct {
	Password string `json:"password"`
}

var _ validation.Validatable = UserDeleteRequest{}

func (r UserDeleteRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Password,
			passwordValidationRules...,
		),
	)
}

// -----------------------------------------------------------------------------
// SeriesCreateRequest
// -----------------------------------------------------------------------------
type SeriesCreateRequest struct {
	Title        string      `json:"title"`
	Descriptions null.String `json:"descriptions"`
	DateStarted  time.Time   `json:"date_started"`
	DateEnded    null.Time   `json:"date_ended"`
}

var _ validation.Validatable = SeriesCreateRequest{}

func (r SeriesCreateRequest) Validate() error {
	timeNow := time.Now()

	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Title,
			validation.Required,
			validation.Length(
				config.Config.Validation.Series.Title.MinLength,
				config.Config.Validation.Series.Title.MaxLength,
			),
		),
		validation.Field(
			&r.Descriptions,
			validation.When(
				r.Descriptions.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.Series.Descriptions.MinLength,
					config.Config.Validation.Series.Descriptions.MaxLength,
				),
			),
		),
		validation.Field(
			&r.DateStarted,
			validation.Required,
			validation.Min(
				time.Date(
					config.Config.Validation.User.Birthdate.MinValue.Year,
					time.Month(
						config.Config.Validation.User.Birthdate.MinValue.Month,
					),
					config.Config.Validation.User.Birthdate.MinValue.Day,
					0, 0, 0, 0, time.UTC,
				),
			),
			validation.Max(
				time.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
					0, 0, 0, 0, time.UTC,
				),
			),
		),
		validation.Field(
			&r.DateEnded,
			validation.When(
				r.DateEnded.Valid,
				validation.Required,
				validation.Min(
					time.Date(
						config.Config.Validation.User.Birthdate.MinValue.Year,
						time.Month(
							config.Config.Validation.User.Birthdate.MinValue.Month,
						),
						config.Config.Validation.User.Birthdate.MinValue.Day,
						0, 0, 0, 0, time.UTC,
					),
				),
				validation.Max(
					time.Date(
						timeNow.Year(), timeNow.Month(), timeNow.Day(),
						0, 0, 0, 0, time.UTC,
					),
				),
			),
		),
	)
}

// -----------------------------------------------------------------------------
// SeriesUpdateRequest
// -----------------------------------------------------------------------------
type SeriesUpdateRequest struct {
	Title        null.String `json:"title"`
	Descriptions null.String `json:"descriptions"`
	DateStarted  null.Time   `json:"date_started"`
	DateEnded    null.Time   `json:"date_ended"`
}

var _ validation.Validatable = SeriesUpdateRequest{}

func (r SeriesUpdateRequest) Validate() error {
	timeNow := time.Now()

	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Title,
			validation.When(
				r.Title.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.Series.Title.MinLength,
					config.Config.Validation.Series.Title.MaxLength,
				),
			),
		),
		validation.Field(
			&r.Descriptions,
			validation.When(
				r.Descriptions.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.Series.Descriptions.MinLength,
					config.Config.Validation.Series.Descriptions.MaxLength,
				),
			),
		),
		validation.Field(
			&r.DateStarted,
			validation.When(
				r.DateStarted.Valid,
				validation.Required,
				validation.Min(
					time.Date(
						config.Config.Validation.User.Birthdate.MinValue.Year,
						time.Month(
							config.Config.Validation.User.Birthdate.MinValue.Month,
						),
						config.Config.Validation.User.Birthdate.MinValue.Day,
						0, 0, 0, 0, time.UTC,
					),
				),
				validation.Max(
					time.Date(
						timeNow.Year(), timeNow.Month(), timeNow.Day(),
						0, 0, 0, 0, time.UTC,
					),
				),
			),
		),
		validation.Field(
			&r.DateEnded,
			validation.When(
				r.DateEnded.Valid,
				validation.Required,
				validation.Min(
					time.Date(
						config.Config.Validation.User.Birthdate.MinValue.Year,
						time.Month(
							config.Config.Validation.User.Birthdate.MinValue.Month,
						),
						config.Config.Validation.User.Birthdate.MinValue.Day,
						0, 0, 0, 0, time.UTC,
					),
				),
				validation.Max(
					time.Date(
						timeNow.Year(), timeNow.Month(), timeNow.Day(),
						0, 0, 0, 0, time.UTC,
					),
				),
			),
		),
	)
}

// -----------------------------------------------------------------------------
// FilmCreateRequest
// -----------------------------------------------------------------------------
type FilmCreateRequest struct {
	Title        string      `json:"title"`
	Descriptions null.String `json:"descriptions"`
	DateReleased time.Time   `json:"date_released"`
	Duration     null.Int    `json:"duration"`
}

var _ validation.Validatable = FilmCreateRequest{}

func (r FilmCreateRequest) Validate() error {
	timeNow := time.Now()

	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Title,
			validation.Required,
			validation.Length(
				config.Config.Validation.Film.Title.MinLength,
				config.Config.Validation.Film.Title.MaxLength,
			),
		),
		validation.Field(
			&r.Descriptions,
			validation.When(
				r.Descriptions.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.Film.Descriptions.MinLength,
					config.Config.Validation.Film.Descriptions.MaxLength,
				),
			),
		),
		validation.Field(
			&r.DateReleased,
			validation.Required,
			validation.Min(
				time.Date(
					config.Config.Validation.User.Birthdate.MinValue.Year,
					time.Month(
						config.Config.Validation.User.Birthdate.MinValue.Month,
					),
					config.Config.Validation.User.Birthdate.MinValue.Day,
					0, 0, 0, 0, time.UTC,
				),
			),
			validation.Max(
				time.Date(
					timeNow.Year(), timeNow.Month(), timeNow.Day(),
					0, 0, 0, 0, time.UTC,
				),
			),
		),
		validation.Field(
			&r.Duration,
			validation.When(
				r.Duration.Valid,
				validation.Required,
				validation.Min(
					config.Config.Validation.Film.Duraion.MinLength,
				),
				validation.Max(
					config.Config.Validation.Film.Duraion.MaxLength,
				),
			),
		),
	)
}

// -----------------------------------------------------------------------------
// FilmUpdateRequest
// -----------------------------------------------------------------------------
type FilmUpdateRequest struct {
	Title        null.String `json:"title"`
	Descriptions null.String `json:"descriptions"`
	DateReleased null.Time   `json:"date_released"`
	Duration     null.Int    `json:"duration"`
}

var _ validation.Validatable = FilmUpdateRequest{}

func (r FilmUpdateRequest) Validate() error {
	timeNow := time.Now()

	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Title,
			validation.When(
				r.Title.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.Film.Title.MinLength,
					config.Config.Validation.Film.Title.MaxLength,
				),
			),
		),
		validation.Field(
			&r.Descriptions,
			validation.When(
				r.Descriptions.Valid,
				validation.Required,
				validation.Length(
					config.Config.Validation.Film.Descriptions.MinLength,
					config.Config.Validation.Film.Descriptions.MaxLength,
				),
			),
		),
		validation.Field(
			&r.DateReleased,
			validation.When(
				r.DateReleased.Valid,
				validation.Required,
				validation.Min(
					time.Date(
						config.Config.Validation.User.Birthdate.MinValue.Year,
						time.Month(
							config.Config.Validation.User.Birthdate.MinValue.Month,
						),
						config.Config.Validation.User.Birthdate.MinValue.Day,
						0, 0, 0, 0, time.UTC,
					),
				),
				validation.Max(
					time.Date(
						timeNow.Year(), timeNow.Month(), timeNow.Day(),
						0, 0, 0, 0, time.UTC,
					),
				),
			),
		),
		validation.Field(
			&r.Duration,
			validation.When(
				r.Duration.Valid,
				validation.Required,
				validation.Min(
					config.Config.Validation.Film.Duraion.MinLength,
				),
				validation.Max(
					config.Config.Validation.Film.Duraion.MaxLength,
				),
			),
		),
	)
}

// -----------------------------------------------------------------------------
// MovieCreateRequest
// -----------------------------------------------------------------------------
type MovieCreateRequest FilmCreateRequest

var _ validation.Validatable = MovieCreateRequest{}

func (r MovieCreateRequest) Validate() error { return FilmCreateRequest(r).Validate() }

// -----------------------------------------------------------------------------
// MovieUpdateRequest
// -----------------------------------------------------------------------------
type MovieUpdateRequest FilmUpdateRequest

var _ validation.Validatable = MovieUpdateRequest{}

func (r MovieUpdateRequest) Validate() error { return FilmUpdateRequest(r).Validate() }

// -----------------------------------------------------------------------------
// EpisodePutRequest
// -----------------------------------------------------------------------------
type EpisodePutRequest FilmCreateRequest

var _ validation.Validatable = EpisodePutRequest{}

func (r EpisodePutRequest) Validate() error { return FilmCreateRequest(r).Validate() }

// -----------------------------------------------------------------------------
// EpisodeUpdateRequest
// -----------------------------------------------------------------------------
type EpisodeUpdateRequest FilmUpdateRequest

var _ validation.Validatable = EpisodeUpdateRequest{}

func (r EpisodeUpdateRequest) Validate() error { return FilmUpdateRequest(r).Validate() }

// -----------------------------------------------------------------------------
// EpisodesPutAllBySeasonRequest
// -----------------------------------------------------------------------------
type EpisodesPutAllBySeasonRequest struct {
	Episodes []*EpisodePutRequest `json:"episodes"`
}

var _ validation.Validatable = EpisodesPutAllBySeasonRequest{}

func (r EpisodesPutAllBySeasonRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Episodes,
			validation.Required,
			validation.Length(
				1,
				config.Config.Validation.Request.Array.MaxLength,
			),
		),
	)
}

// -----------------------------------------------------------------------------
// InvalidationRequest
// -----------------------------------------------------------------------------
type InvalidationRequest struct {
	Invalidation string `json:"invalidation"`
}

var _ validation.Validatable = InvalidationRequest{}

func (r InvalidationRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Invalidation,
			validation.Required,
			validation.Length(
				config.Config.Validation.Request.Invalidation.MinLength,
				config.Config.Validation.Request.Invalidation.MaxLength,
			),
		),
	)
}
