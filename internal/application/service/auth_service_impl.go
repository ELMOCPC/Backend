package service

import (
	"context"
	"encoding/json"

	"github.com/ELMOCPC/Backend/internal/application/usecase"
	cacheRepo "github.com/ELMOCPC/Backend/internal/domain/repository/cache"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/enum"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	repository "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type SignupData struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	TshirtSize    string `json:"tshirt_size"`
	University    string `json:"university"`
	StudentNumber string `json:"student_number"`
	NationalCode  string `json:"national_code"`
}

type AuthService struct {
	userRepository repository.UserRepository
	jwtService     usecase.JwtService
	otpService     usecase.OTPService
	cacheRepo      cacheRepo.CacheRepository
	db             database.Database
}

func NewAuthService(
	userRepository repository.UserRepository,
	jwtService usecase.JwtService,
	otpService usecase.OTPService,
	cacheRepo cacheRepo.CacheRepository,
	db database.Database,
) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		jwtService:     jwtService,
		otpService:     otpService,
		cacheRepo:      cacheRepo,
		db:             db,
	}
}


func (as *AuthService) SignUp(firstName, lastName, email, phone, tshirtSize, university, studentNumber, nationalCode string) error {
	ctx := context.Background()

	// Check if user with this email already exists
	existingUserByEmail, err := as.userRepository.FindUserByEmail(as.db, email)
	if err != nil {
		return err
	}
	if existingUserByEmail != nil {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("email", "already exists")
		return conflictErr
	}

	// Check if user with this phone already exists
	existingUserByPhone, err := as.userRepository.FindUserByPhone(as.db, phone)
	if err != nil {
		return err
	}
	if existingUserByPhone != nil {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("phone", "already exists")
		return conflictErr
	}

	// Store signup details in Redis cache (valid for 10 minutes - same as OTP expiry)
	signupData := SignupData{
		FirstName:     firstName,
		LastName:      lastName,
		Email:         email,
		Phone:         phone,
		TshirtSize:    tshirtSize,
		University:    university,
		StudentNumber: studentNumber,
		NationalCode:  nationalCode,
	}

	if err := as.cacheRepo.Set(ctx, "signup:"+email, signupData, 600); err != nil {
		return err
	}

	// Send OTP to email
	if err := as.otpService.SendOTPEmail(email); err != nil {
		return err
	}

	return nil
}

func (as *AuthService) VerifySignUp(email, code string) (*entity.User, string, string, error) {
	ctx := context.Background()

	// Verify OTP
	if err := as.otpService.ValidateOTP(email, code); err != nil {
		return nil, "", "", err
	}

	// Get signup data from Redis cache
	signupDataJSON, err := as.cacheRepo.Get(ctx, "signup:"+email)
	if err != nil {
		return nil, "", "", err
	}

	if signupDataJSON == "" {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("email", "signup data not found or expired")
		return nil, "", "", validationErr
	}

	var signupData SignupData
	if err := json.Unmarshal([]byte(signupDataJSON), &signupData); err != nil {
		return nil, "", "", err
	}

	// Create new user with signup data
	user := &entity.User{
		FirstName:     signupData.FirstName,
		LastName:      signupData.LastName,
		Email:         email,
		Phone:         signupData.Phone,
		Password:      "", // No password, OTP-based authentication
		TshirtSize:    enum.TshirtSize(signupData.TshirtSize),
		University:    signupData.University,
		StudentNumber: signupData.StudentNumber,
		NationalCode:  signupData.NationalCode,
		EmailVerified: true,
	}

	// Save user to database
	if err := as.userRepository.CreateUser(as.db, user); err != nil {
		return nil, "", "", err
	}

	// Remove from cache
	if err := as.cacheRepo.Delete(ctx, "signup:"+email); err != nil {
		// Log error but don't fail the flow
		_ = err
	}

	// Generate tokens
	accessToken, refreshToken, err := as.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (as *AuthService) Login(email string) error {
	// Find user by email
	user, err := as.userRepository.FindUserByEmail(as.db, email)
	if err != nil {
		return err
	}
	if user == nil {
		return exception.NewUnauthorizedError("user not found", nil)
	}

	// Send OTP to email
	if err := as.otpService.SendOTPEmail(email); err != nil {
		return err
	}

	return nil
}

func (as *AuthService) VerifyLogin(email, code string) (*entity.User, string, string, error) {
	// Verify OTP
	if err := as.otpService.ValidateOTP(email, code); err != nil {
		return nil, "", "", err
	}

	// Find user by email
	user, err := as.userRepository.FindUserByEmail(as.db, email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		return nil, "", "", exception.NewUnauthorizedError("user not found", nil)
	}

	// Generate tokens
	accessToken, refreshToken, err := as.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}
