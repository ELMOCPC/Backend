package postgres

import (
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type OTPRepository interface {
	CreateOTP(db database.Database, otp *entity.OTP) error
	FindOTPByEmail(db database.Database, email string) (*entity.OTP, error)
	UpdateOTP(db database.Database, otp *entity.OTP) error
	DeleteOTP(db database.Database, id uint) error
	DeleteExpiredOTPs(db database.Database) error
}
