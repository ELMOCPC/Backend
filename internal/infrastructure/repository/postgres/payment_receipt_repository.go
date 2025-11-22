package postgres

import (
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	domainRepo "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type PaymentReceiptRepositoryImpl struct{}

func NewPaymentReceiptRepository() domainRepo.PaymentReceiptRepository {
	return &PaymentReceiptRepositoryImpl{}
}

func (repo *PaymentReceiptRepositoryImpl) CreateReceipt(db database.Database, receipt *entity.PaymentReceipt) error {
	result := db.GetDB().Create(receipt)
	return result.Error
}

func (repo *PaymentReceiptRepositoryImpl) FindReceiptByID(db database.Database, id uint) (*entity.PaymentReceipt, error) {
	var receipt entity.PaymentReceipt
	result := db.GetDB().
		Preload("Team").
		Preload("UploadedBy").
		Preload("VerifiedBy").
		First(&receipt, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &receipt, nil
}

func (repo *PaymentReceiptRepositoryImpl) FindReceiptByTeamID(db database.Database, teamID uint) (*entity.PaymentReceipt, error) {
	var receipt entity.PaymentReceipt
	result := db.GetDB().
		Preload("Team").
		Preload("UploadedBy").
		Preload("VerifiedBy").
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		First(&receipt)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &receipt, nil
}

func (repo *PaymentReceiptRepositoryImpl) FindReceiptsByStatus(db database.Database, status entity.ReceiptStatus) ([]*entity.PaymentReceipt, error) {
	var receipts []*entity.PaymentReceipt
	result := db.GetDB().
		Preload("Team").
		Preload("UploadedBy").
		Preload("VerifiedBy").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&receipts)
	return receipts, result.Error
}

func (repo *PaymentReceiptRepositoryImpl) UpdateReceipt(db database.Database, receipt *entity.PaymentReceipt) error {
	result := db.GetDB().Save(receipt)
	return result.Error
}

func (repo *PaymentReceiptRepositoryImpl) DeleteReceipt(db database.Database, id uint) error {
	result := db.GetDB().Delete(&entity.PaymentReceipt{}, id)
	return result.Error
}
