package validator

import (
	"fmt"
	"reflect"

	"github.com/aria3ppp/watchlist-server/internal/modelsfield"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var ErrInvalidFieldOfModel = validation.NewError(
	"validation_model_field_invalid",
	"{{.field}} is not a field of model {{.model}}",
)

type IsValidFieldOfModelRule struct {
	model string
	err   validation.Error
}

func IsValidFieldOfModel(model string) IsValidFieldOfModelRule {
	return IsValidFieldOfModelRule{
		model: model,
		err:   ErrInvalidFieldOfModel,
	}
}

func (r IsValidFieldOfModelRule) Error(message string) IsValidFieldOfModelRule {
	r.err = r.err.SetMessage(message)
	return r
}

func (r IsValidFieldOfModelRule) ErrorObject(
	err validation.Error,
) IsValidFieldOfModelRule {
	r.err = err
	return r
}

func (r IsValidFieldOfModelRule) Validate(value any) error {
	value, isNil := validation.Indirect(value)
	if isNil || validation.IsEmpty(value) {
		return nil
	}

	field, isString := value.(string)
	if !isString {
		return fmt.Errorf(
			"model must be string but it's not: %v",
			reflect.ValueOf(value).Kind(),
		)
	}

	if !modelsfield.Exists(r.model, field) {
		return r.err.SetParams(map[string]any{
			"field": field,
			"model": r.model,
		})
	}

	return nil
}
