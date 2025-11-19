package usecase

import "github.com/ELMOCPC/Backend/internal/domain/entity"

type AuthService interface {
	SignUp(firstName, lastName, email, phone, password string) (*entity.User, string, string, error)
	Login(email, password string) (*entity.User, string, string, error)
}
