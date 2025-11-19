package controller

import (
	"net/http"

	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	"github.com/gin-gonic/gin"
)

type SignUpRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	ID           uint   `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthController struct {
	authService usecase.AuthService
}

func NewAuthController(authService usecase.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ac *AuthController) SignUp(ctx *gin.Context) {
	req := Validated[SignUpRequest](ctx)

	user, accessToken, refreshToken, err := ac.authService.SignUp(
		req.FirstName,
		req.LastName,
		req.Email,
		req.Phone,
		req.Password,
	)
	if err != nil {
		switch err.(type) {
		case *exception.ConflictErrors:
			Response(ctx, http.StatusConflict, "email or phone already exists", nil)
		default:
			panic(err)
		}
		return
	}

	response := &AuthResponse{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Phone:        user.Phone,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	Response(ctx, http.StatusCreated, "user created successfully", response)
}

func (ac *AuthController) Login(ctx *gin.Context) {
	req := Validated[LoginRequest](ctx)

	user, accessToken, refreshToken, err := ac.authService.Login(req.Email, req.Password)
	if err != nil {
		if exception.IsAuthError(err) {
			Response(ctx, http.StatusUnauthorized, "invalid email or password", nil)
		} else {
			panic(err)
		}
		return
	}

	response := &AuthResponse{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Phone:        user.Phone,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	Response(ctx, http.StatusOK, "login successful", response)
}
