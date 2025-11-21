package usecase

import "github.com/ELMOCPC/Backend/internal/domain/entity"

type PaymentReceiptService interface {
	UploadReceipt(teamID, userID uint, receiptImageURL string) (*entity.PaymentReceipt, error)
	VerifyReceipt(receiptID uint, adminID uint, approved bool, notes string) (*entity.PaymentReceipt, error)
	GetPendingReceipts() ([]*entity.PaymentReceipt, error)
	GetReceiptByTeamID(teamID uint) (*entity.PaymentReceipt, error)
}
