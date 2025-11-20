package data

import "kickoff.com/pkg/models"

var NFLTeams = []models.Team{
	// AFC East
	{
		ID: "BUF", Name: "Buffalo Bills", City: "Buffalo", Nickname: "Bills",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCEast,
		PrimaryColor: "#00338D", Founded: 1960, Stadium: "Highmark Stadium",
	},
	{
		ID: "MIA", Name: "Miami Dolphins", City: "Miami", Nickname: "Dolphins",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCEast,
		PrimaryColor: "#008E97", Founded: 1966, Stadium: "Hard Rock Stadium",
	},
	{
		ID: "NE", Name: "New England Patriots", City: "Foxborough", Nickname: "Patriots",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCEast,
		PrimaryColor: "#002244", Founded: 1960, Stadium: "Gillette Stadium",
	},
	{
		ID: "NYJ", Name: "New York Jets", City: "East Rutherford", Nickname: "Jets",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCEast,
		PrimaryColor: "#125740", Founded: 1960, Stadium: "MetLife Stadium",
	},

	// AFC North
	{
		ID: "BAL", Name: "Baltimore Ravens", City: "Baltimore", Nickname: "Ravens",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCNorth,
		PrimaryColor: "#241773", Founded: 1996, Stadium: "M&T Bank Stadium",
	},
	{
		ID: "CIN", Name: "Cincinnati Bengals", City: "Cincinnati", Nickname: "Bengals",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCNorth,
		PrimaryColor: "#FB4F14", Founded: 1968, Stadium: "Paycor Stadium",
	},
	{
		ID: "CLE", Name: "Cleveland Browns", City: "Cleveland", Nickname: "Browns",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCNorth,
		PrimaryColor: "#311D00", Founded: 1946, Stadium: "Cleveland Browns Stadium",
	},
	{
		ID: "PIT", Name: "Pittsburgh Steelers", City: "Pittsburgh", Nickname: "Steelers",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCNorth,
		PrimaryColor: "#FFB612", Founded: 1933, Stadium: "Heinz Field",
	},

	// AFC South
	{
		ID: "HOU", Name: "Houston Texans", City: "Houston", Nickname: "Texans",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCSouth,
		PrimaryColor: "#03202F", Founded: 2002, Stadium: "NRG Stadium",
	},
	{
		ID: "IND", Name: "Indianapolis Colts", City: "Indianapolis", Nickname: "Colts",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCSouth,
		PrimaryColor: "#002C5F", Founded: 1953, Stadium: "Lucas Oil Stadium",
	},
	{
		ID: "JAX", Name: "Jacksonville Jaguars", City: "Jacksonville", Nickname: "Jaguars",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCSouth,
		PrimaryColor: "#006778", Founded: 1995, Stadium: "TIAA Bank Field",
	},
	{
		ID: "TEN", Name: "Tennessee Titans", City: "Nashville", Nickname: "Titans",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCSouth,
		PrimaryColor: "#0C2340", Founded: 1960, Stadium: "Nissan Stadium",
	},

	// AFC West
	{
		ID: "DEN", Name: "Denver Broncos", City: "Denver", Nickname: "Broncos",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCWest,
		PrimaryColor: "#FB4F14", Founded: 1960, Stadium: "Empower Field at Mile High",
	},
	{
		ID: "KC", Name: "Kansas City Chiefs", City: "Kansas City", Nickname: "Chiefs",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCWest,
		PrimaryColor: "#E31837", Founded: 1960, Stadium: "Arrowhead Stadium",
	},
	{
		ID: "LV", Name: "Las Vegas Raiders", City: "Las Vegas", Nickname: "Raiders",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCWest,
		PrimaryColor: "#000000", Founded: 1960, Stadium: "Allegiant Stadium",
	},
	{
		ID: "LAC", Name: "Los Angeles Chargers", City: "Los Angeles", Nickname: "Chargers",
		Conference: models.ConferenceAFC, Division: models.DivisionAFCWest,
		PrimaryColor: "#0080C6", Founded: 1960, Stadium: "SoFi Stadium",
	},

	// NFC East
	{
		ID: "DAL", Name: "Dallas Cowboys", City: "Dallas", Nickname: "Cowboys",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCEast,
		PrimaryColor: "#003594", Founded: 1960, Stadium: "AT&T Stadium",
	},
	{
		ID: "NYG", Name: "New York Giants", City: "East Rutherford", Nickname: "Giants",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCEast,
		PrimaryColor: "#0B2265", Founded: 1925, Stadium: "MetLife Stadium",
	},
	{
		ID: "PHI", Name: "Philadelphia Eagles", City: "Philadelphia", Nickname: "Eagles",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCEast,
		PrimaryColor: "#004C54", Founded: 1933, Stadium: "Lincoln Financial Field",
	},
	{
		ID: "WAS", Name: "Washington Commanders", City: "Landover", Nickname: "Commanders",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCEast,
		PrimaryColor: "#5A1414", Founded: 1932, Stadium: "FedExField",
	},

	// NFC North
	{
		ID: "CHI", Name: "Chicago Bears", City: "Chicago", Nickname: "Bears",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCNorth,
		PrimaryColor: "#0B162A", Founded: 1920, Stadium: "Soldier Field",
	},
	{
		ID: "DET", Name: "Detroit Lions", City: "Detroit", Nickname: "Lions",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCNorth,
		PrimaryColor: "#0076B6", Founded: 1930, Stadium: "Ford Field",
	},
	{
		ID: "GB", Name: "Green Bay Packers", City: "Green Bay", Nickname: "Packers",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCNorth,
		PrimaryColor: "#203731", Founded: 1919, Stadium: "Lambeau Field",
	},
	{
		ID: "MIN", Name: "Minnesota Vikings", City: "Minneapolis", Nickname: "Vikings",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCNorth,
		PrimaryColor: "#4F2683", Founded: 1961, Stadium: "U.S. Bank Stadium",
	},

	// NFC South
	{
		ID: "ATL", Name: "Atlanta Falcons", City: "Atlanta", Nickname: "Falcons",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCSouth,
		PrimaryColor: "#A71930", Founded: 1966, Stadium: "Mercedes-Benz Stadium",
	},
	{
		ID: "CAR", Name: "Carolina Panthers", City: "Charlotte", Nickname: "Panthers",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCSouth,
		PrimaryColor: "#0085CA", Founded: 1995, Stadium: "Bank of America Stadium",
	},
	{
		ID: "NO", Name: "New Orleans Saints", City: "New Orleans", Nickname: "Saints",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCSouth,
		PrimaryColor: "#D3BC8D", Founded: 1967, Stadium: "Caesars Superdome",
	},
	{
		ID: "TB", Name: "Tampa Bay Buccaneers", City: "Tampa", Nickname: "Buccaneers",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCSouth,
		PrimaryColor: "#D50A0A", Founded: 1976, Stadium: "Raymond James Stadium",
	},

	// NFC West
	{
		ID: "ARI", Name: "Arizona Cardinals", City: "Glendale", Nickname: "Cardinals",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCWest,
		PrimaryColor: "#97233F", Founded: 1898, Stadium: "State Farm Stadium",
	},
	{
		ID: "LAR", Name: "Los Angeles Rams", City: "Los Angeles", Nickname: "Rams",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCWest,
		PrimaryColor: "#003594", Founded: 1937, Stadium: "SoFi Stadium",
	},
	{
		ID: "SF", Name: "San Francisco 49ers", City: "San Francisco", Nickname: "49ers",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCWest,
		PrimaryColor: "#AA0000", Founded: 1946, Stadium: "Levi's Stadium",
	},
	{
		ID: "SEA", Name: "Seattle Seahawks", City: "Seattle", Nickname: "Seahawks",
		Conference: models.ConferenceNFC, Division: models.DivisionNFCWest,
		PrimaryColor: "#002244", Founded: 1976, Stadium: "Lumen Field",
	},
}
