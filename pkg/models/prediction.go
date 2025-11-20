package models

import "time"

type PredictionStatus string

const (
	PredictionStatusPending   PredictionStatus = "pending"   // Game hasn't started
	PredictionStatusCorrect   PredictionStatus = "correct"   // Prediction was right
	PredictionStatusIncorrect PredictionStatus = "incorrect" // Prediction was wrong
	PredictionStatusVoid      PredictionStatus = "void"      // Game was canceled/postponed
)

type Prediction struct {
	ID                string           `json:"id"`
	UserID            string           `json:"userId"`
	GameID            string           `json:"gameId"`
	PredictedWinnerID string           `json:"predictedWinnerId"`
	User              *User            `json:"user,omitempty"`            // Populated when needed
	Game              *Game            `json:"game,omitempty"`            // Populated when needed
	PredictedWinner   *Team            `json:"predictedWinner,omitempty"` // Populated when needed
	Status            PredictionStatus `json:"status"`
	Points            int              `json:"points"` // Points earned (1 for correct, 0 for incorrect)
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
}

type CreatePredictionRequest struct {
	UserID            string `json:"userId"`
	GameID            string `json:"gameId"`
	PredictedWinnerID string `json:"predictedWinnerId"`
}

type PredictionsResponse struct {
	Predictions []Prediction `json:"predictions"`
	Total       int          `json:"total"`
}

type UserPredictionsResponse struct {
	UserID      string       `json:"userId"`
	User        *User        `json:"user,omitempty"`
	Predictions []Prediction `json:"predictions"`
	Total       int          `json:"total"`
	Correct     int          `json:"correct"`
	Incorrect   int          `json:"incorrect"`
	Pending     int          `json:"pending"`
	Percentage  float64      `json:"percentage"`
}
