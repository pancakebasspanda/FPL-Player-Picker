package storage

type PlayerData struct {
	PlayerSummaryData PlayersSummaryData
	PlayerStatsData   PlayerStatsData
}

type PlayersSummaryData struct {
	PlayerInfo PlayerInfoCol
	Cols       []Col
}

func (p *PlayersSummaryData) getPlayerCols(playerInfo PlayerInfoCol) []Col {
	if p.PlayerInfo.Name == playerInfo.Name &&
		p.PlayerInfo.Team == playerInfo.Team &&
		p.PlayerInfo.Position == playerInfo.Position {
		return p.Cols

	}
	return []Col{}
}

type PlayerStatsData struct {
	PlayerInfo PlayerInfoCol
	Rows       []Row
}

type Row struct {
	Cols []Col
}

type Col struct {
	Name  string
	Value string
}

type PlayerInfoCol struct {
	Name     string
	Team     string
	Position string
}
