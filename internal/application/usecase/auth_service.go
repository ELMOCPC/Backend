package usecase

import "github.com/ELMOCPC/Backend/internal/domain/entity"

type AuthService interface {
	SignUp(firstName, lastName, email, phone, tshirtSize, university, studentNumber string) error
	VerifySignUp(email, code string) (*entity.User, string, string, error)
	Login(email string) error
	VerifyLogin(email, code string) (*entity.User, string, string, error)
}
