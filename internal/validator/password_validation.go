package validator

import (
	"fmt"
	"reflect"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var ErrPasswordInvalid = validation.NewError(
	"validation_password_invalid",
	"password must contain at least "+
		"{{.num}} numbers, "+
		"{{.lower}} lowercase letters, "+
		"{{.upper}} uppercase letters, "+
		"{{.special}} special characters",
)

const specialChars = "@#$%^&*_-+=(){}[]]|\\:;\"'`<>.,~!?/"

func Password(
	numbers, lowerLetters, upperLetters, specialChars int,
) PasswordRule {
	return PasswordRule{
		numbers:      numbers,
		lowerLetters: lowerLetters,
		upperLetters: upperLetters,
		specialChars: specialChars,
	}
}

type PasswordRule struct {
	numbers, lowerLetters, upperLetters, specialChars int
}

func (r PasswordRule) Validate(value any) error {
	value, isNil := validation.Indirect(value)
	if isNil || validation.IsEmpty(value) {
		return nil
	}

	str, isString := value.(string)
	if !isString {
		return fmt.Errorf(
			"password validation is not operable on value %v",
			reflect.ValueOf(value).Kind(),
		)
	}

	var nums, lowers, uppers, specials int
	for _, ch := range str {
		if '0' <= ch && ch <= '9' {
			nums++
			continue
		}
		if 'a' <= ch && ch <= 'z' {
			lowers++
			continue
		}
		if 'A' <= ch && ch <= 'Z' {
			uppers++
			continue
		}
		if strings.ContainsRune(specialChars, ch) {
			specials++
			continue
		}
	}

	if nums < r.numbers || lowers < r.lowerLetters || uppers < r.upperLetters ||
		specials < r.specialChars {
		return ErrPasswordInvalid.SetParams(map[string]any{
			"num":     r.numbers,
			"lower":   r.lowerLetters,
			"upper":   r.upperLetters,
			"special": r.specialChars,
		})
	}

	return nil
}
