package entity

import (
	"github.com/ELMOCPC/Backend/internal/domain/enum"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type Permission struct {
	database.Model
	Name     string                  `gorm:"not null"`
	Type     enum.PermissionType     `gorm:"not null"`
	Category enum.PermissionCategory `gorm:"not null"`
	Roles    []Role                  `gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE"`
}
