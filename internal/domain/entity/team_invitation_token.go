package entity

import (
	"time"

	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type TeamInvitationToken struct {
	database.Model
	TeamID        uint       `gorm:"index"`
	InvitedByID   uint       `gorm:"index"`
	Email         string     `gorm:"type:varchar(255);index"`
	Token         string     `gorm:"type:varchar(255);unique;index"`
	FirstName     string
	LastName      string
	Phone         string
	TshirtSize    string
	University    string
	StudentNumber string
	ExpiresAt     time.Time
	UsedAt        *time.Time
	Accepted      bool `gorm:"default:false"`
	Rejected      bool `gorm:"default:false"`
}
