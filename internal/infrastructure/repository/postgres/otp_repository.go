package postgres

import (
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	domainRepo "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type OTPRepositoryImpl struct{}

func NewOTPRepository() domainRepo.OTPRepository {
	return &OTPRepositoryImpl{}
}

func (repo *OTPRepositoryImpl) CreateOTP(db database.Database, otp *entity.OTP) error {
	result := db.GetDB().Create(otp)
	return result.Error
}

func (repo *OTPRepositoryImpl) FindOTPByEmail(db database.Database, email string) (*entity.OTP, error) {
	var otp entity.OTP
	result := db.GetDB().
		Where("email = ? AND used = ? AND expires_at > NOW()", email, false).
		Order("created_at DESC").
		First(&otp)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &otp, nil
}

func (repo *OTPRepositoryImpl) UpdateOTP(db database.Database, otp *entity.OTP) error {
	result := db.GetDB().Save(otp)
	return result.Error
}

func (repo *OTPRepositoryImpl) DeleteOTP(db database.Database, id uint) error {
	result := db.GetDB().Delete(&entity.OTP{}, id)
	return result.Error
}

func (repo *OTPRepositoryImpl) DeleteExpiredOTPs(db database.Database) error {
	result := db.GetDB().Where("expires_at < NOW()").Delete(&entity.OTP{})
	return result.Error
}
