package entity

import (
	"github.com/ELMOCPC/Backend/internal/domain/enum"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type User struct {
	database.Model
	FirstName     string            `gorm:"type:varchar(255)"`
	LastName      string            `gorm:"type:varchar(255)"`
	Email         string            `gorm:"type:varchar(255);unique"`
	Phone         string            `gorm:"type:varchar(11);unique"`
	Password      string            `gorm:"type:varchar(255)"`
	TshirtSize    enum.TshirtSize   `gorm:"type:varchar(10)"`
	University    string            `gorm:"type:varchar(255)"`
	StudentNumber string            `gorm:"type:varchar(50)"`
	EmailVerified bool              `gorm:"default:false"`
	Roles         []Role            `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE"`
	TeamID        *uint             `gorm:"index"`
}
