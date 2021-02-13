package storage

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"pancakebasspanda/fpl_player_picker/model"
	"strconv"
	"strings"
)

const (
	_query = `
INSERT INTO player_stats_summary(name, position, team, form, cost, selected_percentage, total_points )
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name)
DO UPDATE SET
position = excluded.position,
team = excluded.team,
form = excluded.form,
cost = excluded.cost,
selected_percentage = excluded.selected_percentage,
total_points = excluded.total_points

`
)

func (s SqlLite) SavePlayerSummaries(data PlayersSummaryData) {

	// convert to database types

	summaryItem, ok := convertToPlayerSummary(data)

	if !ok {
		return
	}

	statement, _ := s.db.Prepare(_query)

	// insert
	result, err := statement.Exec(summaryItem.Player.Name,
		summaryItem.Player.Position,
		summaryItem.Player.Team,
		summaryItem.Form,
		summaryItem.Cost,
		summaryItem.SelectedPercentage,
		summaryItem.Points)

	if err != nil {
		log.WithError(err).WithError(err).WithError(err).Fatal("exec insert")
	}

	defer statement.Close()

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		log.WithError(err).WithError(err).WithError(err).Fatal("rows affected")
	}

	log.WithField("inserted row", fmt.Sprintf("%d", rowsAffected)).Info("inserted fields")

	return
}

func convertToPlayerSummary(data PlayersSummaryData) (model.PlayerSummary, bool) {
	var summaryItem model.PlayerSummary
	ok := true
	summaryItem.Player.Name = data.PlayerInfo.Name
	summaryItem.Player.Position = data.PlayerInfo.Position
	summaryItem.Player.Team = data.PlayerInfo.Team

	for _, col := range data.Cols {
		if value, ok := fieldMapping[col.Name]; ok {
			switch strings.ToLower(value) {
			case "cost":
				f, err := strconv.ParseFloat(col.Value, 64)

				if err != nil {
					ok = false
					log.WithError(err).Error("cost")
				}
				summaryItem.Cost = f
			case "selected_percentage":

				summaryItem.SelectedPercentage = col.Value
			case "form":
				f, err := strconv.ParseFloat(col.Value, 64)

				if err != nil {
					ok = false
					log.WithError(err).Error("form")
				}
				summaryItem.Form = f
			case "points":
				i, err := strconv.Atoi(col.Value)

				if err != nil {
					ok = false
					log.WithError(err).Error("points")
				}
				summaryItem.Points = i
			}
		}
	}

	if !ok {
		log.WithField("player", data.PlayerInfo.Name).Error("malformed summary data for player")

		return summaryItem, false
	}

	return summaryItem, true
}
