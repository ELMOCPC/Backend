package controller

import (
	"net/http"

	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	"github.com/ELMOCPC/Backend/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService usecase.AuthService
	otpService  usecase.OTPService
}

func NewAuthController(authService usecase.AuthService, otpService usecase.OTPService) *AuthController {
	return &AuthController{
		authService: authService,
		otpService:  otpService,
	}
}


func (ac *AuthController) Login(ctx *gin.Context) {
	req := Validated[dto.LoginRequest](ctx)

	if err := ac.authService.Login(req.Email); err != nil {
		if exception.IsAuthError(err) {
			Response(ctx, http.StatusUnauthorized, "user not found", nil)
		} else {
			panic(err)
		}
		return
	}

	Response(ctx, http.StatusOK, "OTP sent to your email", nil)
}

func (ac *AuthController) VerifyLogin(ctx *gin.Context) {
	req := Validated[dto.VerifyOTPRequest](ctx)

	user, accessToken, refreshToken, err := ac.authService.VerifyLogin(req.Email, req.Code)
	if err != nil {
		switch err.(type) {
		case exception.NotFoundError:
			Response(ctx, http.StatusNotFound, "OTP not found or expired", nil)
		case exception.ValidationErrors:
			Response(ctx, http.StatusBadRequest, "invalid or expired OTP", nil)
		default:
			panic(err)
		}
		return
	}

	response := &dto.AuthResponse{
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

func (ac *AuthController) SignUp(ctx *gin.Context) {
	req := Validated[dto.CompleteSignUpRequest](ctx)

	if err := ac.authService.SignUp(
		req.FirstName,
		req.LastName,
		req.Email,
		req.Phone,
		req.TshirtSize,
		req.University,
		req.StudentNumber,
	); err != nil {
		switch err.(type) {
		case *exception.ConflictErrors:
			Response(ctx, http.StatusConflict, "email or phone already exists", nil)
		default:
			panic(err)
		}
		return
	}

	Response(ctx, http.StatusOK, "OTP sent to your email", nil)
}

func (ac *AuthController) VerifySignUp(ctx *gin.Context) {
	req := Validated[dto.VerifyOTPRequest](ctx)

	user, accessToken, refreshToken, err := ac.authService.VerifySignUp(req.Email, req.Code)
	if err != nil {
		switch err.(type) {
		case exception.NotFoundError:
			Response(ctx, http.StatusNotFound, "OTP not found or expired", nil)
		case exception.ValidationErrors:
			Response(ctx, http.StatusBadRequest, "invalid or expired OTP", nil)
		default:
			panic(err)
		}
		return
	}

	response := &dto.AuthResponse{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Phone:        user.Phone,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	Response(ctx, http.StatusCreated, "account created successfully", response)
}
