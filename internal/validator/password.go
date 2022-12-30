package validator

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var ErrInvalidPassword = validation.NewError(
	"validation_password_invalid",
	"password must contain at least "+
		"{{.num}} numbers, "+
		"{{.lower}} lowercase letters, "+
		"{{.upper}} uppercase letters, "+
		"{{.special}} special characters",
)

type IsPasswordRule struct {
	numbers      int
	lowerLetters int
	upperLetters int
	specialChars int

	err validation.Error
}

func IsPassword() IsPasswordRule {
	return IsPasswordRule{
		err: ErrInvalidPassword,
	}
}

func (r IsPasswordRule) Numbers(n int) IsPasswordRule {
	r.numbers = n
	return r
}

func (r IsPasswordRule) LowerLetters(n int) IsPasswordRule {
	r.lowerLetters = n
	return r
}

func (r IsPasswordRule) UpperLetters(n int) IsPasswordRule {
	r.upperLetters = n
	return r
}

func (r IsPasswordRule) SpecialChars(n int) IsPasswordRule {
	r.specialChars = n
	return r
}

func (r IsPasswordRule) Error(message string) IsPasswordRule {
	r.err = r.err.SetMessage(message)
	return r
}

func (r IsPasswordRule) ErrorObject(err validation.Error) IsPasswordRule {
	r.err = err
	return r
}

func (r IsPasswordRule) Validate(value any) error {
	value, isNil := validation.Indirect(value)
	if isNil /*|| validation.IsEmpty(value)*/ {
		return nil
	}

	password, err := validation.EnsureString(value)
	if err != nil {
		return err
	}

	var nums, lowers, uppers, specials int
	const specialChars = "@#$%^&*_-+=(){}[]]|\\:;\"'`<>.,~!?/"

	for _, ch := range password {
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

	if r.numbers > 0 && nums < r.numbers ||
		r.lowerLetters > 0 && lowers < r.lowerLetters ||
		r.upperLetters > 0 && uppers < r.upperLetters ||
		r.specialChars > 0 && specials < r.specialChars {
		return r.err.SetParams(map[string]any{
			"num":     r.numbers,
			"lower":   r.lowerLetters,
			"upper":   r.upperLetters,
			"special": r.specialChars,
		})
	}

	return nil
}
