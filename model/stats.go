package model

type PlayerSummary struct {
	Player             Player
	Cost               float64 `sql:"cost"`
	Form               float64 `sql:"form"`
	Points             int     `sql:"points"`
	SelectedPercentage string  `sql:"selected_percentage"`
}

type PlayerStat struct {
	Player          Player
	Opposition      string `sql:"opposition"`
	GameWeek        int    `sql:"game_week"`
	Points          int    `sql:"points"`
	MinutesPlayed   int    `sql:"minutes_played"`
	GoalsScored     int    `sql:"goals_scored"`
	Assists         int    `sql:"assists"`
	CleanSheets     int    `sql:"clean_sheets"`
	GoalsConceded   int    `sql:"goals_conceded"`
	GoalsSaved      int    `sql:"goals_saved"`
	OwnGoals        int    `sql:"own_goals"`
	PenaltiesSaved  int    `sql:"penalties_saved"`
	PenaltiesMissed int    `sql:"penalties_missed"`
	YellowCards     int    `sql:"yellow_cards"`
	RedCard         int    `sql:"red_cards"`
	Bonus           int    `sql:"bonus"`
}

type Player struct {
	Name     string `sql:"name"`
	Team     string `sql:"team"`
	Position string `sql:"position"`
}
