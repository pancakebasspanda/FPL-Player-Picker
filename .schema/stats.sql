CREATE TABLE `player_stats_summary` (
                                        name TEXT NOT NULL,
                                        team TEXT NOT NULL,
                                        position TEXT NOT NULL,
                                        cost  REAL NULL,
                                        selected_percentage TEXT NULL,
                                        form REAL NULL,
                                        total_points INTEGER NULL,
                                        PRIMARY KEY (name)
);

CREATE UNIQUE INDEX idx_form
ON player_stats_summary (form);

CREATE UNIQUE INDEX idx_position
ON player_stats_summary (position );


CREATE TABLE `player_stats_weekly` (
                                       name TEXT NOT NULL,
                                       team TEXT NOT NULL,
                                       position TEXT NOT NULL,
                                       opposition TEXT NULL,
                                       game_week  INTEGER NULL,
                                       points  INTEGER NULL,
                                       minutes_played  INTEGER NULL,
                                       goals_scored  INTEGER NULL,
                                       assists INTEGER NULL,
                                       clean_sheets INTEGER NULL,
                                       goals_conceded INTEGER NULL,
                                       goals_saved INTEGER NULL,
                                       own_goals INTEGER NULL,
                                       penalties_saved INTEGER NULL,
                                       penalties_missed INTEGER NULL,
                                       yellow_cards INTEGER NULL,
                                       red_cards INTEGER NULL,
                                       bonus INTEGER NULL,
                                       PRIMARY KEY (name)
);


CREATE UNIQUE INDEX idx_game_week
ON player_stats_weekly (game_week);