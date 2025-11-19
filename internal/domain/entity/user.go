package entity

import "github.com/ELMOCPC/Backend/internal/infrastructure/database"

type User struct {
	database.Model
	FirstName string `gorm:"type:varchar(255)"`
	LastName  string `gorm:"type:varchar(255)"`
	Email     string `gorm:"type:varchar(255);unique"`
	Phone     string `gorm:"type:varchar(11);unique"`
	Password  string `gorm:"type:varchar(255)"`
	Roles     []Role `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE"`
	TeamID    *uint  `gorm:"index"`
	Team      *Team  `gorm:"foreignKey:TeamID;constraint:OnDelete:SET NULL"`
}
