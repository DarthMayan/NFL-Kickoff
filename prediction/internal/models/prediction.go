package models

import (
	"time"
	"gorm.io/gorm"
)

type PredictionStatus string

const (
	PredictionStatusPending   PredictionStatus = "pending"
	PredictionStatusCorrect   PredictionStatus = "correct"
	PredictionStatusIncorrect PredictionStatus = "incorrect"
	PredictionStatusVoid      PredictionStatus = "void"
)

// Prediction representa una predicción de un usuario sobre un juego
type Prediction struct {
	ID                string           `gorm:"primaryKey;type:varchar(50)" json:"id"`
	UserID            string           `gorm:"not null;type:varchar(50);index" json:"userId"`
	GameID            string           `gorm:"not null;type:varchar(50);index" json:"gameId"`
	PredictedWinnerID string           `gorm:"not null;type:varchar(10)" json:"predictedWinnerId"`
	Status            PredictionStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Points            int              `gorm:"default:0" json:"points"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName especifica el nombre de la tabla
func (Prediction) TableName() string {
	return "predictions"
}
