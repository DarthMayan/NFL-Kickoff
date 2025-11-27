package models

import (
	"time"
	"gorm.io/gorm"
)

type GameStatus string

const (
	GameStatusScheduled GameStatus = "scheduled"
	GameStatusLive      GameStatus = "live"
	GameStatusCompleted GameStatus = "completed"
	GameStatusPostponed GameStatus = "postponed"
	GameStatusCanceled  GameStatus = "canceled"
)

// Game representa un juego NFL
type Game struct {
	ID           string         `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Week         int            `gorm:"not null" json:"week"`
	Season       int            `gorm:"not null;default:2024" json:"season"`
	HomeTeamID   string         `gorm:"not null;type:varchar(10);index" json:"homeTeamId"`
	AwayTeamID   string         `gorm:"not null;type:varchar(10);index" json:"awayTeamId"`
	GameTime     time.Time      `gorm:"not null" json:"gameTime"`
	Status       GameStatus     `gorm:"type:varchar(20);default:'scheduled'" json:"status"`
	HomeScore    int            `gorm:"default:0" json:"homeScore"`
	AwayScore    int            `gorm:"default:0" json:"awayScore"`
	WinnerTeamID string         `gorm:"type:varchar(10)" json:"winnerTeamId,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica el nombre de la tabla
func (Game) TableName() string {
	return "games"
}

// Conference representa la conferencia NFL
type Conference string

const (
	ConferenceAFC Conference = "AFC"
	ConferenceNFC Conference = "NFC"
)

// Division representa la división NFL
type Division string

const (
	DivisionAFCEast  Division = "AFC East"
	DivisionAFCNorth Division = "AFC North"
	DivisionAFCSouth Division = "AFC South"
	DivisionAFCWest  Division = "AFC West"
	DivisionNFCEast  Division = "NFC East"
	DivisionNFCNorth Division = "NFC North"
	DivisionNFCSouth Division = "NFC South"
	DivisionNFCWest  Division = "NFC West"
)

// Team representa un equipo NFL
type Team struct {
	ID         string     `gorm:"primaryKey;type:varchar(10)" json:"id"`
	Name       string     `gorm:"not null;type:varchar(100)" json:"name"`
	City       string     `gorm:"not null;type:varchar(100)" json:"city"`
	Conference Conference `gorm:"type:varchar(3);not null" json:"conference"`
	Division   Division   `gorm:"type:varchar(20);not null" json:"division"`
	LogoURL    string     `gorm:"type:varchar(500)" json:"logoUrl"`
	Stadium    string     `gorm:"type:varchar(200)" json:"stadium"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName especifica el nombre de la tabla
func (Team) TableName() string {
	return "teams"
}
