package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ELMOCPC/Backend/bootstrap"
	"github.com/ELMOCPC/Backend/internal/application/usecase"
	cacheRepo "github.com/ELMOCPC/Backend/internal/domain/repository/cache"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
)

type OTPData struct {
	Code      string `json:"code"`
	Attempts  int    `json:"attempts"`
	Used      bool   `json:"used"`
	ExpiresAt int64  `json:"expires_at"`
}

type OTPServiceImpl struct {
	cacheRepo    cacheRepo.CacheRepository
	emailService usecase.EmailService
	env          *bootstrap.Env
}

func NewOTPService(
	cacheRepo cacheRepo.CacheRepository,
	emailService usecase.EmailService,
	env *bootstrap.Env,
) usecase.OTPService {
	return &OTPServiceImpl{
		cacheRepo:    cacheRepo,
		emailService: emailService,
		env:          env,
	}
}

func (os *OTPServiceImpl) GenerateOTP(email string) (string, error) {
	ctx := context.Background()

	// Generate a random OTP code
	otp := os.generateRandomCode(os.env.OTP.Length)

	// Create OTP data
	otpData := OTPData{
		Code:      otp,
		Attempts:  0,
		Used:      false,
		ExpiresAt: time.Now().Add(time.Duration(os.env.OTP.ExpiryMinute) * time.Minute).Unix(),
	}

	// Save OTP to Redis cache
	if err := os.cacheRepo.Set(ctx, "otp:"+email, otpData, os.env.OTP.ExpiryMinute*60); err != nil {
		return "", err
	}

	return otp, nil
}

func (os *OTPServiceImpl) ValidateOTP(email, code string) error {
	ctx := context.Background()

	// Get OTP from Redis cache
	otpJSON, err := os.cacheRepo.Get(ctx, "otp:"+email)
	if err != nil {
		return err
	}

	if otpJSON == "" {
		return exception.NotFoundError{Item: "OTP"}
	}

	var otpData OTPData
	if err := json.Unmarshal([]byte(otpJSON), &otpData); err != nil {
		return err
	}

	// Check if OTP is expired
	if time.Now().Unix() > otpData.ExpiresAt {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("otp", "expired")
		return validationErr
	}

	// Check if OTP code matches
	if otpData.Code != code {
		otpData.Attempts++
		if otpData.Attempts >= os.env.OTP.MaxAttempts {
			otpData.Used = true
		}
		// Update in Redis
		if err := os.cacheRepo.Set(ctx, "otp:"+email, otpData, os.env.OTP.ExpiryMinute*60); err != nil {
			return err
		}
		validationErr := exception.ValidationErrors{}
		validationErr.Add("otp", "invalid")
		return validationErr
	}

	// Mark OTP as used and delete from cache
	if err := os.cacheRepo.Delete(ctx, "otp:"+email); err != nil {
		// Log error but don't fail
		_ = err
	}

	return nil
}

func (os *OTPServiceImpl) SendOTPEmail(email string) error {
	// Generate OTP
	otp, err := os.GenerateOTP(email)
	if err != nil {
		return err
	}

	// Send email with OTP
	subject := "Your ELMOCPC Verification Code"
	body := fmt.Sprintf(`
Dear User,

Your ELMOCPC verification code is: %s

This code will expire in %d minutes.

If you did not request this code, please ignore this email.

Best regards,
ELMOCPC Team
`, otp, os.env.OTP.ExpiryMinute)

	// For now, we'll use console output if SMTP is not configured
	if err := os.emailService.SendEmail(email, subject, body); err != nil {
		return err
	}

	return nil
}

func (os *OTPServiceImpl) generateRandomCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = digits[rand.Intn(len(digits))]
	}
	return string(code)
}

var _ usecase.OTPService = (*OTPServiceImpl)(nil)
