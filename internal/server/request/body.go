package request

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type TokenBody struct {
	Token string `json:"token"`
}

var _ validation.Validatable = TokenBody{}

func (r TokenBody) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(
			&r.Token,
			validation.Required,
			is.UUID,
		),
	)
}
