package main

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"pancakebasspanda/fpl_player_picker/app"
	"pancakebasspanda/fpl_player_picker/storage"
	"time"
)

var (
	_dataSource string
	_logLevel   string
)

func init() {
	flag.StringVar(&_dataSource, "data-source", "db/premier_league_football", "sqlite file path")
	flag.StringVar(&_logLevel, "log-level", "info", "log level")
}

func main() {
	flag.Parse()

	lvl, err := log.ParseLevel(_logLevel)

	if err != nil {
		log.WithError(err).Fatal("log level")
	}

	log.SetLevel(lvl)

	db, err := sql.Open("sqlite3", _dataSource)

	if err != nil {
		log.WithError(err).WithError(err).Fatal("connecting to db")
	}

	defer db.Close()

	store := storage.New(db)

	start := time.Now()

	app.Runner(store)

	elapsed := time.Since(start)

	log.Infof("scraper took %s", elapsed)

}
