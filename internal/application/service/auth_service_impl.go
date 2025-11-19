package service

import (
	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	repository "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepository repository.UserRepository
	jwtService     usecase.JwtService
	db             database.Database
}

func NewAuthService(
	userRepository repository.UserRepository,
	jwtService usecase.JwtService,
	db database.Database,
) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		jwtService:     jwtService,
		db:             db,
	}
}

func (as *AuthService) SignUp(firstName, lastName, email, phone, password string) (*entity.User, string, string, error) {
	// Check if user with this email already exists
	existingUserByEmail, err := as.userRepository.FindUserByEmail(as.db, email)
	if err != nil {
		return nil, "", "", err
	}
	if existingUserByEmail != nil {
		conflictErr := &exception.ConflictErrors{}
		conflictErr.Add("email", "already exists")
		return nil, "", "", conflictErr
	}

	// Check if user with this phone already exists
	existingUserByPhone, err := as.userRepository.FindUserByPhone(as.db, phone)
	if err != nil {
		return nil, "", "", err
	}
	if existingUserByPhone != nil {
		conflictErr := &exception.ConflictErrors{}
		conflictErr.Add("phone", "already exists")
		return nil, "", "", conflictErr
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	// Create new user
	user := &entity.User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
		Password:  string(hashedPassword),
	}

	// Save user to database
	if err := as.userRepository.CreateUser(as.db, user); err != nil {
		return nil, "", "", err
	}

	// Generate tokens
	accessToken, refreshToken, err := as.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (as *AuthService) Login(email, password string) (*entity.User, string, string, error) {
	// Find user by email
	user, err := as.userRepository.FindUserByEmail(as.db, email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		return nil, "", "", exception.NewUnauthorizedError("invalid email or password", nil)
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", exception.NewUnauthorizedError("invalid email or password", nil)
	}

	// Generate tokens
	accessToken, refreshToken, err := as.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}
