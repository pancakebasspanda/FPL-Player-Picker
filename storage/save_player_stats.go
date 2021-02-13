package storage

import log "github.com/sirupsen/logrus"

func (s *SqlLite) SavePlayerStats() {

	// insert
	stmt, err := s.db.Prepare("insert into player_stats_summary values ('Kane', 'TOT', 'FWD', 11.2, '46.6', 7.8, 142) ON CONFLICT DO NOTHING;")
	if err != nil {
		log.WithError(err).WithError(err).WithError(err).Fatal("prepare insert")
	}

	res, err := stmt.Exec()
	if err != nil {
		log.WithError(err).WithError(err).Fatal("insert to player summary")
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.WithError(err).WithError(err).Fatal("insert id")
	}

	log.WithError(err).WithError(err).Info(id)

	return

}
