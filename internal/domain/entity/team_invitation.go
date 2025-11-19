package entity

import "github.com/ELMOCPC/Backend/internal/infrastructure/database"

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRejected InvitationStatus = "rejected"
)

type TeamInvitation struct {
	database.Model
	TeamID       uint               `gorm:"not null;index"`
	Team         *Team              `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
	Email        string             `gorm:"type:varchar(255);not null;index"`
	Status       InvitationStatus   `gorm:"type:varchar(50);default:'pending'"`
	InvitedByID  uint               `gorm:"not null"`
	InvitedBy    *User              `gorm:"foreignKey:InvitedByID;constraint:OnDelete:SET NULL"`
}

func (TeamInvitation) TableName() string {
	return "team_invitations"
}
