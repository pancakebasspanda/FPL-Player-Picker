package scraper

import (
	"github.com/mxschmitt/playwright-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"pancakebasspanda/fpl_player_picker/storage"

	"context"
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

const (
	_pageURL = "https://fantasy.premierleague.com/statistics"
)

var (
	_maxWorkers = runtime.GOMAXPROCS(0)
	_sem        = semaphore.NewWeighted(int64(_maxWorkers / 2))
	_wg         sync.WaitGroup
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

func (s *scraper) findMaxPagination() (int, error) {
	page, err := s.browser.NewPage()

	defer page.Close()

	if err != nil {
		return 0, fmt.Errorf("new page: %w", err)
	}

	if _, err := page.Goto(_pageURL, playwright.PageGotoOptions{Timeout: playwright.Int(10000)}); err != nil {
		return 0, fmt.Errorf("go to page: %w", err)
	}

	pagesText, err := page.InnerText("#root table + div div:nth-of-type(1)")

	if err != nil {
		return 0, fmt.Errorf("selecting max pages element: %w", err)
	}

	maxPagesStr := pagesText[len(pagesText)-2:]

	log.Info("max pages: " + maxPagesStr)

	maxPages, err := strconv.Atoi(maxPagesStr)

	if err != nil {
		return 0, fmt.Errorf("parsing max pagination number: %w", err)
	}

	return maxPages, nil

}

func (s *scraper) ScrapPage(ctx context.Context) error {
	maxPages, err := s.findMaxPagination()
	if err != nil {
		return err
	}

	for p := 1; p <= maxPages; p++ {
		page, err := s.browser.NewPage()

		if err != nil {
			log.WithError(err).Error("new page")
		}

		if _, err := page.Goto(_pageURL, playwright.PageGotoOptions{Timeout: playwright.Int(10000)}); err != nil {
			log.WithError(err).Error("go to page")
		}

		for nextClick := 1; nextClick < p; nextClick++ {
			err = page.Click("#root table + div button:nth-of-type(3)")

			if err != nil {
				log.WithError(err).Error("pagination click")
			}
		}

		page.WaitForLoadState("load")

		// get current pagination number
		pageNo, err := page.InnerText("#root strong")

		if err != nil {
			log.WithError(err).Error("selecting max pages dix")
		}

		log.WithField("page no", pageNo).Info("scraping....")

		_wg.Add(1)
		_sem.Acquire(ctx, 1)
		go s.scapePlayerSummaryPage(page)

	}

	_wg.Wait()

	log.Info("finished scraping page")

	return nil
}

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

	//var wg sync.WaitGroup
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

				//wg.Add(1)

				//go scrapePlayerStatsPage(rootDialog, &wg, playerStatsChan, player)
				// TODO os.Remove()
				log.WithField("player", player).Info("scraping player: ")
				closePlayerDialog(rootDialog)
				//wg.Done()
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

	//wg.Wait()

	//close(playerStatsChan)

	return

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
