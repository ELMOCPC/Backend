package entity

import "github.com/ELMOCPC/Backend/internal/infrastructure/database"

type TeamStatus string

const (
	TeamStatusDraft      TeamStatus = "draft"
	TeamStatusSubmitted  TeamStatus = "submitted"
	TeamStatusValidated  TeamStatus = "validated"
	TeamStatusRejected   TeamStatus = "rejected"
)

type Team struct {
	database.Model
	Name        string                `gorm:"type:varchar(255);not null"`
	Description string                `gorm:"type:text"`
	CreatorID   uint                  `gorm:"not null;index"`
	Creator     *User                 `gorm:"foreignKey:CreatorID;constraint:OnDelete:SET NULL"`
	Status      TeamStatus            `gorm:"type:varchar(50);default:'draft'"`
	Members     []User                `gorm:"many2many:team_members;constraint:OnDelete:CASCADE"`
	Invitations []*TeamInvitation     `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
}

func (Team) TableName() string {
	return "teams"
}
