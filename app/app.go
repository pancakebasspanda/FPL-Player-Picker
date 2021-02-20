package app

import (
	"context"
	_ "github.com/mattn/go-sqlite3"
	"pancakebasspanda/fpl_player_picker/storage"

	"github.com/mxschmitt/playwright-go"
	log "github.com/sirupsen/logrus"
	"pancakebasspanda/fpl_player_picker/scraper"
)

func Runner(storage storage.Storage) {

	ctx := context.Background()

	pw, err := playwright.Run()

	if err != nil {
		log.WithError(err).Fatal("running playwright")
	}

	browser, err := pw.Chromium.Launch()

	if err != nil {
		log.WithError(err).Fatal("launching browser")
	}

	s := scraper.New(browser, storage)

	if err := s.ScrapPage(ctx); err != nil {
		log.WithError(err).Fatal("scraping page")
	}

	if err := browser.Close(); err != nil {
		log.WithError(err).Fatal("closing page")
	}

	if err = pw.Stop(); err != nil {
		log.WithError(err).Fatal("stopping playwright")
	}

}
