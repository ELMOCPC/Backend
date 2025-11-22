package entity

import (
	"time"

	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type OTP struct {
	database.Model
	Email     string    `gorm:"type:varchar(255);index"`
	Code      string    `gorm:"type:varchar(255)"`
	ExpiresAt time.Time
	Attempts  int `gorm:"default:0"`
	Used      bool `gorm:"default:false"`
}
