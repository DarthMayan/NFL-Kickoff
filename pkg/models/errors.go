package models

import "errors"

var (
	// User errors
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidUserData = errors.New("invalid user data")

	// Team errors
	ErrTeamNotFound    = errors.New("team not found")
	ErrInvalidTeamData = errors.New("invalid team data")

	// Game errors
	ErrGameNotFound       = errors.New("game not found")
	ErrGameAlreadyStarted = errors.New("game already started")
	ErrGameCompleted      = errors.New("game already completed")
	ErrInvalidGameData    = errors.New("invalid game data")
	ErrTeamNotInGame      = errors.New("team not participating in this game")

	// Prediction errors
	ErrPredictionNotFound    = errors.New("prediction not found")
	ErrPredictionExists      = errors.New("prediction already exists for this game")
	ErrPredictionTooLate     = errors.New("cannot predict after game has started")
	ErrInvalidPredictionData = errors.New("invalid prediction data")

	// General errors
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrUnauthorized       = errors.New("unauthorized")
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
