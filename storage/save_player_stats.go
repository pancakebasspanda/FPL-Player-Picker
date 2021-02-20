package storage

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"pancakebasspanda/fpl_player_picker/model"
)

func (s *SqlLite) SavePlayerStats(data PlayerStatsData) {

	// convert to database types
	statItems := convertToPlayerStatsData(data)

	statement, _ := s.db.Prepare(_query)
	defer statement.Close()

	for _, statItem := range statItems {

		// insert
		result, err := statement.Exec(statItem.Player.Name,
			statItem.Player.Position,
			statItem.Player.Team,
			statItem.GameWeek,
			statItem.Points,
			statItem.MinutesPlayed,
			statItem.GoalsScored,
			statItem.Assists,
			statItem.CleanSheets,
			statItem.GoalsConceded,
			statItem.OwnGoals,
			statItem.PenaltiesSaved,
			statItem.PenaltiesMissed,
			statItem.YellowCards,
			statItem.RedCard,
			statItem.Saves,
			statItem.Bonus)

		if err != nil {
			log.WithError(err).WithError(err).Error("exec insert")

			continue
		}

		rowsAffected, err := result.RowsAffected()

		if err != nil {
			log.WithError(err).WithError(err).Error("rows affected")

			continue
		}

		log.WithField("inserted row", fmt.Sprintf("%d", rowsAffected)).
			WithField("statItem", statItem).
			Debug("inserted fields")

	}

	return

}

func convertToPlayerStatsData(data PlayerStatsData) []model.PlayerStat {
	statItems := make([]model.PlayerStat, 0)

	var statItem model.PlayerStat

	statItem.Player.Name = data.PlayerInfo.Name
	statItem.Player.Position = data.PlayerInfo.Position
	statItem.Player.Team = data.PlayerInfo.Team

	for _, row := range data.Rows {
		for _, col := range row.Cols {
			if value, ok := fieldMapping[col.Name]; ok {
				switch strings.ToLower(value) {
				case "game_week":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("game_week", i).WithError(err).Error("game_week")
						continue
					}

					statItem.GameWeek = i
				case "opposition":
					statItem.Opposition = col.Value
				case "minutes_played":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("minutes_played", i).WithError(err).Error("minutes_played")
						continue
					}

					statItem.MinutesPlayed = i
				case "goals_scored":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("goals_scored", i).WithError(err).Error("goals_scored")
						continue
					}

					statItem.GoalsScored = i
				case "goals_conceded":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("goals_conceded", i).WithError(err).Error("goals_conceded")
						continue
					}

					statItem.GoalsConceded = i
				case "assists":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("assists", i).WithError(err).Error("assists")
						continue
					}

					statItem.Assists = i
				case "clean_sheets":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("clean_sheets", i).WithError(err).Error("clean_sheets")
						continue
					}

					statItem.CleanSheets = i
				case "own_goals":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("own_goals", i).WithError(err).Error("own_goals")
						continue
					}

					statItem.OwnGoals = i
				case "penalties_saved":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("penalties_saved", i).WithError(err).Error("penalties_saved")
						continue
					}

					statItem.PenaltiesSaved = i
				case "penalties_missed":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("penalties_missed", i).WithError(err).Error("penalties_missed")
						continue
					}

					statItem.PenaltiesMissed = i
				case "yellow_cards":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("yellow_cards", i).WithError(err).Error("yellow_cards")
						continue
					}

					statItem.YellowCards = i
				case "red_cards":
					i, err := strconv.Atoi(col.Value)

					if err != nil {

						log.WithField("red_cards", i).WithError(err).Error("red_cards")
					}

					statItem.RedCard = i
				case "points":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("points", i).WithError(err).Error("points")
						continue
					}
					statItem.Points = i
				case "bonus":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("bonus", i).WithError(err).Error("bonus")
						continue
					}
					statItem.Bonus = i
				case "saves":
					i, err := strconv.Atoi(col.Value)

					if err != nil {
						log.WithField("saves", i).WithError(err).Error("saves")
						continue
					}
					statItem.Saves = i
				}
			}
		}

		statItems = append(statItems, statItem)
	}

	return statItems
}
