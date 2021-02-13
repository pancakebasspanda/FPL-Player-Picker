package scraper

import (
	"github.com/mxschmitt/playwright-go"
	log "github.com/sirupsen/logrus"
	"pancakebasspanda/fpl_player_picker/storage"
	"sync"
)

const (
	_pageURL = "https://fantasy.premierleague.com/statistics"
)

type scraper struct {
	browser playwright.Browser
	storage storage.Storage
}

func New(browser playwright.Browser, storage storage.Storage) *scraper {
	return &scraper{
		browser: browser,
		storage: storage,
	}
}

func (s *scraper) ScrapPage() error {
	i := 0
Exit:
	for {
		page, err := s.browser.NewPage()

		if err != nil {
			log.Fatalf("could not create page: %v", err)
		}

		if _, err = page.Goto(_pageURL, playwright.PageGotoOptions{Timeout: playwright.Int(10000)}); err != nil {
			log.Error("could not go to url: %v", err)
			continue
		}

		i++

		for nextClick := 0; nextClick < 1; nextClick++ {

			err = page.Click("div:nth-child(2) div  div.Layout__Main-eg6k6r-1.haICgV  div.sc-AykKC.sc-AykKD.iKIfJP button:nth-child(4)")
			if err != nil {
				log.Error(err)
				break Exit
			}
		}

		page.WaitForLoadState("load")

		log.Infof("scraping page %d", i)

		go s.scapePlayerSummaryPage(page)

	}

	log.Infof("successfully finished scraping page")

	return nil
}

func (s *scraper) scapePlayerSummaryPage(page playwright.Page) {

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
	//playerStatsChan := make(chan storage.PlayerStatsData)

	// read the player stats after we scrap summaries
	//go func() {
	//	for statChan := range playerStatsChan {
	//
	//		log.Infof("player : %+v", statChan.PlayerInfo)
	//		for _, playerStat := range statChan.Rows {
	//
	//			log.Info(playerStat.Cols)
	//
	//			log.Infoln()
	//		}
	//	}
	//}()

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
			log.Errorf("could not select ColData: %v", err)
		}

		player := storage.PlayerInfoCol{}

		for i, col := range cols {

			switch i {
			case 0:

				// click button to open player page
				button, err := col.QuerySelector("button")

				if err != nil {
					log.WithError(err).Error("error retrieving player information error")
				}

				err = button.Click(playwright.ElementHandleClickOptions{})

				if err != nil {
					log.WithError(err).Error("error clicking player information error")
				}

				rootDialog, err := page.WaitForSelector("#root-dialog", playwright.PageWaitForSelectorOptions{
					State: playwright.String("attached"),
				})

				if err != nil {
					log.WithError(err).Error("error retrieving iFrame with player info")
				}

				// player name
				name, err := rootDialog.QuerySelector("h2")

				if err != nil {
					log.Errorf("could not player name element: %v", err)
				}

				player.Name, err = name.InnerText()

				if err != nil {
					log.Errorf("could not player name text: %v", err)
				}

				// player position
				position, err := rootDialog.QuerySelector("h2 + span")

				if err != nil {
					log.Errorf("could not get Name player postiion element: %v", err)
				}

				player.Position, err = position.InnerText()

				if err != nil {
					log.Errorf("could not player position text: %v", err)
				}

				// player team
				team, err := rootDialog.QuerySelector("h2 + span + div")

				if err != nil {
					log.Errorf("could not get Name player team element: %v", err)
				}

				player.Team, err = team.InnerText()

				if err != nil {
					log.Errorf("could not player team text: %v", err)
				}

				if err != nil {
					log.Errorf("could not get span element: %v", err)
				}

				playerSummaries[irow].PlayerInfo = player

				wg.Add(1)

				//go scrapePlayerStatsPage(rootDialog, &wg, playerStatsChan, player)
				// TODO os.Remove()
				closePlayerDialog(rootDialog)
				wg.Done()
			default:

				colText, err := col.InnerText()

				if err != nil {
					log.Errorf("error getting column tableData: %v", err)
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

	//close(playerStatsChan)

}

func scrapePlayerStatsPage(frame playwright.ElementHandle, wg *sync.WaitGroup, statsChan chan storage.PlayerStatsData, player storage.PlayerInfoCol) {
	defer wg.Done()

	// column colNames
	colNames := make([]storage.Col, 0)
	tableHead, err := frame.QuerySelector("thead")

	if err != nil {
		log.Errorf("could not get entries: %v", err)
	}

	cols, err := tableHead.QuerySelectorAll("th") //only one tr since header

	for _, col := range cols {
		colText, err := col.InnerText()

		if err != nil {
			log.Errorf("error getting column tableData: %v", err)
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
		log.Errorf("could not get entries: %v", err)
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
			log.Errorf("player stats col selector: %v", err)
		}

		for i, col := range cols {

			colText, err := col.InnerText()

			if err != nil {
				log.Errorf("error getting player stats col text: %v", err)
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
