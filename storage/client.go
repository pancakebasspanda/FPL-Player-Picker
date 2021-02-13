package storage

import (
	"database/sql"
)

type Storage interface {
	SavePlayerStats()
	SavePlayerSummaries(data PlayersSummaryData)
}

type SqlLite struct {
	db *sql.DB
}

func New(database *sql.DB) Storage {
	return &SqlLite{db: database}
}
