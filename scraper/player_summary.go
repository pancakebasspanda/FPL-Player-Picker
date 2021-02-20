package scraper

import (
	"github.com/mxschmitt/playwright-go"
	log "github.com/sirupsen/logrus"
	"pancakebasspanda/fpl_player_picker/storage"
	"sync"
)

func (s *scraper) scapePlayerSummaryPage(page playwright.Page) {
	defer page.Close()
	defer _sem.Release(1)
	defer _wg.Done()

	// summaries data
	playerSummaries := make([]storage.PlayersSummaryData, 0)

	root, err := page.QuerySelector("#root")
	if err != nil {
		log.Fatalf("could not get entries: %v", err)
	}

	tableHead, err := root.QuerySelector("thead")
	if err != nil {
		log.Errorf("could not get entries: %v", err)
	}

	cols, err := tableHead.QuerySelectorAll("th")

	colNames := make([]storage.Col, 0)

	for i, col := range cols {

		switch i {
		case 0:
			// skip as there is no col Name
			continue
		default:
			colValue, err := col.InnerText()

			if err != nil {
				log.Fatalf("error getting column tableData: %v", err)
			}

			// storing player in its on struct so skip col 1
			if i != 1 {
				colNames = append(colNames, storage.Col{
					Name: colValue,
				})
			}
		}

	}

	tableBody, err := root.QuerySelector("tbody")

	if err != nil {
		log.Errorf("could not get entries: %v", err)
	}

	rows, err := tableBody.QuerySelectorAll("tr")

	var wg sync.WaitGroup
	playerStatsChan := make(chan storage.PlayerStatsData)

	// read the player stats after we scrap summaries
	go func() {
		for playerStats := range playerStatsChan {
			log.WithField("player", playerStats.PlayerInfo.Name).Debug("sawing player stats")

			s.storage.SavePlayerStats(playerStats)
		}

	}()

	for irow, row := range rows {

		// need cols for all player summaries
		// copy the col Names across (Go points to original struct if we use assignment)
		rowHeader := make([]storage.Col, 0)
		for _, col := range colNames {
			rowHeader = append(rowHeader, storage.Col{
				Name: col.Name,
			})
		}

		// need row for all player summaries
		playerSummaries = append(playerSummaries, storage.PlayersSummaryData{
			Cols: rowHeader,
		})

		cols, err := row.QuerySelectorAll("td")

		if err != nil {
			log.WithError(err).Error("columns selector")
		}

		for i, col := range cols {
			switch i {
			case 0:

				// wait for the player stats page to load
				rootDialog, err := page.WaitForSelector("#root-dialog", playwright.PageWaitForSelectorOptions{
					State: playwright.String("attached"),
				})

				player, err := getPlayerData(col, rootDialog)

				if err != nil {
					continue
				}

				wg.Add(1)

				log.WithField("player", player.Name).Debug("scraping player stats data")

				go scrapePlayerStatsPage(page, rootDialog, &wg, playerStatsChan, player)

			default:

				colText, err := col.InnerText()

				if err != nil {
					log.WithError(err).Error("col text")
				}

				// skip 1st column as we get player details from player stats page
				// index of columns starts from 3rd col as first 2 are i button and player col
				if i != 1 {
					playerSummaries[irow].Cols[i-2].Value = colText
				}

			}

		}
	}

	for _, row := range playerSummaries {

		s.storage.SavePlayerSummaries(row)
	}

	wg.Wait()

	close(playerStatsChan)

	return

}
