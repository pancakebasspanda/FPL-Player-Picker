package scraper

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"sync"

	"github.com/mxschmitt/playwright-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"pancakebasspanda/fpl_player_picker/storage"
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

	if _, err := page.Goto(_pageURL); err != nil {
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

		if _, err := page.Goto(_pageURL); err != nil {
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
			log.WithError(err).Error("selecting max pages div")
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
