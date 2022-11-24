package hasher

import "golang.org/x/crypto/bcrypt"

//go:generate mockgen -destination mock_hasher/mock_hasher.go . Interface

type Interface interface {
	GenerateFromPassword(password []byte, cost int) ([]byte, error)
	CompareHashAndPassword(hashedPassword, password []byte) error
}

func NewBcrypt() Bcrypt {
	return Bcrypt{}
}

type Bcrypt struct{}

var _ Interface = (Bcrypt{})

func (a Bcrypt) GenerateFromPassword(
	password []byte,
	cost int,
) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, cost)
}

func (a Bcrypt) CompareHashAndPassword(
	hashedPassword, password []byte,
) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return ErrMismatchedHashAndPassword
		}
		return err
	}
	return nil
}
