package models

type Conference string
type Division string

const (
	ConferenceAFC Conference = "AFC"
	ConferenceNFC Conference = "NFC"
)

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

type Team struct {
	ID           string     `json:"id"`           // Short code like "KC", "SF"
	Name         string     `json:"name"`         // Full name like "Kansas City Chiefs"
	City         string     `json:"city"`         // City name like "Kansas City"
	Nickname     string     `json:"nickname"`     // Team nickname like "Chiefs"
	Conference   Conference `json:"conference"`   // AFC or NFC
	Division     Division   `json:"division"`     // Division name
	PrimaryColor string     `json:"primaryColor"` // Hex color code
	LogoURL      string     `json:"logoUrl"`      // URL to team logo
	Founded      int        `json:"founded"`      // Year founded
	Stadium      string     `json:"stadium"`      // Stadium name
}

type TeamsResponse struct {
	Teams []Team `json:"teams"`
	Total int    `json:"total"`
}
