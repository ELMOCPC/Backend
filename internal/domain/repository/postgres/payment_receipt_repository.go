package postgres

import (
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type PaymentReceiptRepository interface {
	CreateReceipt(db database.Database, receipt *entity.PaymentReceipt) error
	FindReceiptByID(db database.Database, id uint) (*entity.PaymentReceipt, error)
	FindReceiptByTeamID(db database.Database, teamID uint) (*entity.PaymentReceipt, error)
	FindReceiptsByStatus(db database.Database, status entity.ReceiptStatus) ([]*entity.PaymentReceipt, error)
	UpdateReceipt(db database.Database, receipt *entity.PaymentReceipt) error
	DeleteReceipt(db database.Database, id uint) error
}
