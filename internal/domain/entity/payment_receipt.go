package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type ReceiptStatus string

const (
	ReceiptStatusPending  ReceiptStatus = "pending"
	ReceiptStatusApproved ReceiptStatus = "approved"
	ReceiptStatusRejected ReceiptStatus = "rejected"
)

type PaymentReceipt struct {
	database.Model
	TeamID          uint          `gorm:"index;not null"`
	Team            *Team         `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
	UploadedByID    uint          `gorm:"not null"`
	UploadedBy      *User         `gorm:"foreignKey:UploadedByID;constraint:OnDelete:SET NULL"`
	ReceiptImageURL string        `gorm:"type:varchar(500)"`
	Status          ReceiptStatus `gorm:"type:varchar(50);default:'pending'"`
	VerifiedByID    *uint         `gorm:"index"`
	VerifiedBy      *User         `gorm:"foreignKey:VerifiedByID;constraint:OnDelete:SET NULL"`
	VerificationNotes DataMap      `gorm:"type:jsonb;serializer:json"`
	VerifiedAt      *time.Time
}

type DataMap map[string]interface{}

func (d DataMap) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DataMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(bytes, &d)
}
