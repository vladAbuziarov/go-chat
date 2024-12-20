package hash

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrCannotHashPassword = errors.New("cannot hash password: %s")
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Join(fmt.Errorf(ErrCannotHashPassword.Error(), password), err)
	}
	return string(hash), nil
}

func (s *Service) CompareHashWithPassword(hash, password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("failed to process password hash: %w", err)
	}
	return true, nil
}
