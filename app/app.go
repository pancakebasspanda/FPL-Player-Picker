package app

import (
	_ "github.com/mattn/go-sqlite3"
	"pancakebasspanda/fpl_player_picker/storage"

	"github.com/mxschmitt/playwright-go"
	log "github.com/sirupsen/logrus"
	"pancakebasspanda/fpl_player_picker/scraper"
)

const (
	_pageURL = "https://fantasy.premierleague.com/statistics"
)

type app struct {
}

func New() *app {
	return &app{}

}

func (a *app) Runner(storage storage.Storage) {

	pw, err := playwright.Run()

	if err != nil {
		log.WithError(err).Fatalf("running playwright")
	}

	browser, err := pw.Chromium.Launch()

	if err != nil {
		log.Error(err)
	}

	defer browser.Close()

	s := scraper.New(browser, storage)

	if err := s.ScrapPage(); err != nil {
		log.WithError(err).Error("error scraping webpage")
	}

	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}

}
