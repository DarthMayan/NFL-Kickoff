package models

import "time"

type GameStatus string

const (
	GameStatusScheduled GameStatus = "scheduled"
	GameStatusLive      GameStatus = "live"
	GameStatusCompleted GameStatus = "completed"
	GameStatusPostponed GameStatus = "postponed"
	GameStatusCanceled  GameStatus = "canceled"
)

type Game struct {
	ID           string     `json:"id"`
	Week         int        `json:"week"`
	Season       int        `json:"season"`
	HomeTeamID   string     `json:"homeTeamId"`
	AwayTeamID   string     `json:"awayTeamId"`
	HomeTeam     *Team      `json:"homeTeam,omitempty"` // Populated when needed
	AwayTeam     *Team      `json:"awayTeam,omitempty"` // Populated when needed
	GameTime     time.Time  `json:"gameTime"`
	Status       GameStatus `json:"status"`
	HomeScore    int        `json:"homeScore"`
	AwayScore    int        `json:"awayScore"`
	WinnerTeamID string     `json:"winnerTeamId,omitempty"` // Set when game is completed
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type CreateGameRequest struct {
	Week       int       `json:"week"`
	Season     int       `json:"season"`
	HomeTeamID string    `json:"homeTeamId"`
	AwayTeamID string    `json:"awayTeamId"`
	GameTime   time.Time `json:"gameTime"`
}

type UpdateGameRequest struct {
	GameTime  *time.Time  `json:"gameTime,omitempty"`
	Status    *GameStatus `json:"status,omitempty"`
	HomeScore *int        `json:"homeScore,omitempty"`
	AwayScore *int        `json:"awayScore,omitempty"`
}

type GamesResponse struct {
	Games []Game `json:"games"`
	Total int    `json:"total"`
}

type GamesByWeekResponse struct {
	Week   int    `json:"week"`
	Season int    `json:"season"`
	Games  []Game `json:"games"`
	Total  int    `json:"total"`
}
