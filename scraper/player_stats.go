package scraper

import (
	"github.com/mxschmitt/playwright-go"
	log "github.com/sirupsen/logrus"
	"pancakebasspanda/fpl_player_picker/storage"
	"sync"
)

func scrapePlayerStatsPage(page playwright.Page, frame playwright.ElementHandle, wg *sync.WaitGroup, statsChan chan storage.PlayerStatsData, player storage.PlayerInfoCol) {
	defer wg.Done()

	// column colNames
	colNames := make([]storage.Col, 0)

	// TODO investigate why the look up fails if we do it outside of this method. page seems not to be thread safe
	_, err := page.WaitForSelector("#root-dialog tbody", playwright.PageWaitForSelectorOptions{
		State: playwright.String("attached"),
	})

	if err != nil {
		log.WithError(err).Error("player stats body not rendered")
	}

	tableHead, err := frame.QuerySelector("thead")

	if err != nil {
		log.WithError(err).Error("player stats headers")
	}

	cols, err := tableHead.QuerySelectorAll("th") //only one tr since header

	for _, col := range cols {
		colText, err := col.InnerText()

		if err != nil {
			log.WithError(err).Error("player stats header col")
		}

		if colText == "Totals" {
			break
		}

		colNames = append(colNames, storage.Col{
			Name: colText,
		})

	}

	tableBody, err := frame.QuerySelector("tbody")

	if err != nil {
		log.WithError(err).Error("table body selector")
	}

	rows, err := tableBody.QuerySelectorAll("tr")

	playerStatsData := storage.PlayerStatsData{
		PlayerInfo: player,
	}

	playerRowData := make([]storage.Row, 0)

	for irow, row := range rows {

		// copy the col Names across (Go points to original struct if we use assignment)
		rowHeader := make([]storage.Col, 0)
		for _, col := range colNames {
			rowHeader = append(rowHeader, storage.Col{
				Name: col.Name,
			})
		}

		// need row for all player summaries
		playerRowData = append(playerRowData, storage.Row{
			Cols: rowHeader,
		})

		cols, err := row.QuerySelectorAll("td")

		if err != nil {
			log.WithError(err).Error("player stats cols")
		}

		for i, col := range cols {

			colText, err := col.InnerText()

			if err != nil {
				log.WithError(err).Error("player stats text")
			}

			if colText == "Totals" {
				playerRowData = playerRowData[:len(playerRowData)-1]
				break
			}

			playerRowData[irow].Cols[i].Value = colText
		}

	}

	playerStatsData.Rows = playerRowData

	// need to hit the close button on the summary data
	closePlayerDialog(frame)

	statsChan <- playerStatsData
}

func closePlayerDialog(frame playwright.ElementHandle) {

	buttons, err := frame.QuerySelectorAll("button")

	if err != nil {
		log.WithError(err).Error("player stats page close button")
	}

	err = buttons[0].Click()

	if err != nil {
		log.WithError(err).Error("click player stats page close button")
	}

}

func getPlayerData(col playwright.ElementHandle, dialog playwright.ElementHandle) (storage.PlayerInfoCol, error) {
	player := storage.PlayerInfoCol{}

	// click button to open player page
	button, err := col.QuerySelector("button")

	if err != nil {
		log.WithError(err).Error("error retrieving player information error")

		return player, err
	}

	err = button.Click(playwright.ElementHandleClickOptions{})

	if err != nil {
		log.WithError(err).Error("error clicking player information error")

		return player, err
	}

	if err != nil {
		log.WithError(err).Error("error retrieving iFrame with player info")

		return player, err
	}

	// player name
	name, err := dialog.QuerySelector("h2")

	if err != nil {
		log.WithError(err).Error("player name selector")

		return player, err
	}

	player.Name, err = name.InnerText()

	if err != nil {
		log.WithError(err).Error("player name text")

		return player, err
	}

	// player position
	position, err := dialog.QuerySelector("h2 + span")

	if err != nil {
		log.WithError(err).Error("player position selector")

		return player, err
	}

	player.Position, err = position.InnerText()

	if err != nil {
		log.WithError(err).Error("player position text")

		return player, err
	}

	// player team
	team, err := dialog.QuerySelector("h2 + span + div")

	if err != nil {
		log.WithError(err).Error("player team selector")

		return player, err
	}

	player.Team, err = team.InnerText()

	if err != nil {
		log.WithError(err).Error("player team text")

		return player, err
	}

	return player, nil

}
