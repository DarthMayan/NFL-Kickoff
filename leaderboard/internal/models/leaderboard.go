package models

import (
	"time"
	"gorm.io/gorm"
)

// UserStats representa las estadísticas de un usuario en el leaderboard
type UserStats struct {
	ID                 string         `gorm:"primaryKey;type:varchar(50)" json:"id"`
	UserID             string         `gorm:"uniqueIndex;not null;type:varchar(50)" json:"userId"`
	TotalPredictions   int            `gorm:"default:0" json:"totalPredictions"`
	CorrectPredictions int            `gorm:"default:0" json:"correctPredictions"`
	WrongPredictions   int            `gorm:"default:0" json:"wrongPredictions"`
	TotalPoints        int            `gorm:"default:0" json:"totalPoints"`
	Rank               int            `gorm:"default:0" json:"rank"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica el nombre de la tabla
func (UserStats) TableName() string {
	return "user_stats"
}
