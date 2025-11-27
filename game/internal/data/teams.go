package data

import "kickoff.com/game/internal/models"

var NFLTeams = []models.Team{
	// AFC East
	{ID: "BUF", Name: "Buffalo Bills", City: "Buffalo", Conference: models.ConferenceAFC, Division: models.DivisionAFCEast, Stadium: "Highmark Stadium"},
	{ID: "MIA", Name: "Miami Dolphins", City: "Miami", Conference: models.ConferenceAFC, Division: models.DivisionAFCEast, Stadium: "Hard Rock Stadium"},
	{ID: "NE", Name: "New England Patriots", City: "Foxborough", Conference: models.ConferenceAFC, Division: models.DivisionAFCEast, Stadium: "Gillette Stadium"},
	{ID: "NYJ", Name: "New York Jets", City: "East Rutherford", Conference: models.ConferenceAFC, Division: models.DivisionAFCEast, Stadium: "MetLife Stadium"},

	// AFC North
	{ID: "BAL", Name: "Baltimore Ravens", City: "Baltimore", Conference: models.ConferenceAFC, Division: models.DivisionAFCNorth, Stadium: "M&T Bank Stadium"},
	{ID: "CIN", Name: "Cincinnati Bengals", City: "Cincinnati", Conference: models.ConferenceAFC, Division: models.DivisionAFCNorth, Stadium: "Paycor Stadium"},
	{ID: "CLE", Name: "Cleveland Browns", City: "Cleveland", Conference: models.ConferenceAFC, Division: models.DivisionAFCNorth, Stadium: "Cleveland Browns Stadium"},
	{ID: "PIT", Name: "Pittsburgh Steelers", City: "Pittsburgh", Conference: models.ConferenceAFC, Division: models.DivisionAFCNorth, Stadium: "Heinz Field"},

	// AFC South
	{ID: "HOU", Name: "Houston Texans", City: "Houston", Conference: models.ConferenceAFC, Division: models.DivisionAFCSouth, Stadium: "NRG Stadium"},
	{ID: "IND", Name: "Indianapolis Colts", City: "Indianapolis", Conference: models.ConferenceAFC, Division: models.DivisionAFCSouth, Stadium: "Lucas Oil Stadium"},
	{ID: "JAX", Name: "Jacksonville Jaguars", City: "Jacksonville", Conference: models.ConferenceAFC, Division: models.DivisionAFCSouth, Stadium: "TIAA Bank Field"},
	{ID: "TEN", Name: "Tennessee Titans", City: "Nashville", Conference: models.ConferenceAFC, Division: models.DivisionAFCSouth, Stadium: "Nissan Stadium"},

	// AFC West
	{ID: "DEN", Name: "Denver Broncos", City: "Denver", Conference: models.ConferenceAFC, Division: models.DivisionAFCWest, Stadium: "Empower Field at Mile High"},
	{ID: "KC", Name: "Kansas City Chiefs", City: "Kansas City", Conference: models.ConferenceAFC, Division: models.DivisionAFCWest, Stadium: "Arrowhead Stadium"},
	{ID: "LV", Name: "Las Vegas Raiders", City: "Las Vegas", Conference: models.ConferenceAFC, Division: models.DivisionAFCWest, Stadium: "Allegiant Stadium"},
	{ID: "LAC", Name: "Los Angeles Chargers", City: "Los Angeles", Conference: models.ConferenceAFC, Division: models.DivisionAFCWest, Stadium: "SoFi Stadium"},

	// NFC East
	{ID: "DAL", Name: "Dallas Cowboys", City: "Dallas", Conference: models.ConferenceNFC, Division: models.DivisionNFCEast, Stadium: "AT&T Stadium"},
	{ID: "NYG", Name: "New York Giants", City: "East Rutherford", Conference: models.ConferenceNFC, Division: models.DivisionNFCEast, Stadium: "MetLife Stadium"},
	{ID: "PHI", Name: "Philadelphia Eagles", City: "Philadelphia", Conference: models.ConferenceNFC, Division: models.DivisionNFCEast, Stadium: "Lincoln Financial Field"},
	{ID: "WAS", Name: "Washington Commanders", City: "Landover", Conference: models.ConferenceNFC, Division: models.DivisionNFCEast, Stadium: "FedExField"},

	// NFC North
	{ID: "CHI", Name: "Chicago Bears", City: "Chicago", Conference: models.ConferenceNFC, Division: models.DivisionNFCNorth, Stadium: "Soldier Field"},
	{ID: "DET", Name: "Detroit Lions", City: "Detroit", Conference: models.ConferenceNFC, Division: models.DivisionNFCNorth, Stadium: "Ford Field"},
	{ID: "GB", Name: "Green Bay Packers", City: "Green Bay", Conference: models.ConferenceNFC, Division: models.DivisionNFCNorth, Stadium: "Lambeau Field"},
	{ID: "MIN", Name: "Minnesota Vikings", City: "Minneapolis", Conference: models.ConferenceNFC, Division: models.DivisionNFCNorth, Stadium: "U.S. Bank Stadium"},

	// NFC South
	{ID: "ATL", Name: "Atlanta Falcons", City: "Atlanta", Conference: models.ConferenceNFC, Division: models.DivisionNFCSouth, Stadium: "Mercedes-Benz Stadium"},
	{ID: "CAR", Name: "Carolina Panthers", City: "Charlotte", Conference: models.ConferenceNFC, Division: models.DivisionNFCSouth, Stadium: "Bank of America Stadium"},
	{ID: "NO", Name: "New Orleans Saints", City: "New Orleans", Conference: models.ConferenceNFC, Division: models.DivisionNFCSouth, Stadium: "Caesars Superdome"},
	{ID: "TB", Name: "Tampa Bay Buccaneers", City: "Tampa", Conference: models.ConferenceNFC, Division: models.DivisionNFCSouth, Stadium: "Raymond James Stadium"},

	// NFC West
	{ID: "ARI", Name: "Arizona Cardinals", City: "Glendale", Conference: models.ConferenceNFC, Division: models.DivisionNFCWest, Stadium: "State Farm Stadium"},
	{ID: "LAR", Name: "Los Angeles Rams", City: "Los Angeles", Conference: models.ConferenceNFC, Division: models.DivisionNFCWest, Stadium: "SoFi Stadium"},
	{ID: "SF", Name: "San Francisco 49ers", City: "San Francisco", Conference: models.ConferenceNFC, Division: models.DivisionNFCWest, Stadium: "Levi's Stadium"},
	{ID: "SEA", Name: "Seattle Seahawks", City: "Seattle", Conference: models.ConferenceNFC, Division: models.DivisionNFCWest, Stadium: "Lumen Field"},
}
