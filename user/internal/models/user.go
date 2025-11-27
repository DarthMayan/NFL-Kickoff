package models

import (
	"time"
	"gorm.io/gorm"
)

// User representa un usuario en el sistema
type User struct {
	ID        string         `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null;type:varchar(100)" json:"username"`
	Email     string         `gorm:"uniqueIndex;not null;type:varchar(255)" json:"email"`
	FullName  string         `gorm:"type:varchar(255)" json:"fullName"`
	Active    bool           `gorm:"default:true;not null" json:"active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica el nombre de la tabla
func (User) TableName() string {
	return "users"
}
