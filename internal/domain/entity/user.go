package entity

import "github.com/ELMOCPC/Backend/internal/infrastructure/database"

type User struct {
	database.Model
	Roles []Role `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE"`
}
