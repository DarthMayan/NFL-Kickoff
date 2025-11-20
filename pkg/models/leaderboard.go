package models

type UserStats struct {
	UserID       string  `json:"userId"`
	User         *User   `json:"user,omitempty"`
	Rank         int     `json:"rank"`
	TotalPicks   int     `json:"totalPicks"`
	CorrectPicks int     `json:"correctPicks"`
	Percentage   float64 `json:"percentage"`
	Points       int     `json:"points"`
	Week         int     `json:"week,omitempty"`   // For weekly leaderboards
	Season       int     `json:"season,omitempty"` // For season leaderboards
}

type LeaderboardResponse struct {
	Leaderboard   []UserStats `json:"leaderboard"`
	TotalUsers    int         `json:"totalUsers"`
	Week          int         `json:"week,omitempty"`
	Season        int         `json:"season,omitempty"`
	GamesFinished int         `json:"gamesFinished"`
	GamesTotal    int         `json:"gamesTotal"`
	LastUpdated   string      `json:"lastUpdated"`
}

type WeeklyLeaderboard struct {
	Week        int         `json:"week"`
	Season      int         `json:"season"`
	Leaderboard []UserStats `json:"leaderboard"`
	TotalUsers  int         `json:"totalUsers"`
}

type SeasonLeaderboard struct {
	Season      int         `json:"season"`
	Leaderboard []UserStats `json:"leaderboard"`
	TotalUsers  int         `json:"totalUsers"`
}

type UserStatsDetail struct {
	UserStats         UserStats    `json:"userStats"`
	Predictions       []Prediction `json:"predictions"`
	RecentPerformance []UserStats  `json:"recentPerformance"` // Last 4 weeks
}
