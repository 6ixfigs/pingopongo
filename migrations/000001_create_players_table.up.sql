CREATE TABLE players (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	user_id varchar(255) NOT NULL,
	channel_id varchar(255) NOT NULL,
	team_id varchar(255) NOT NULL,
	full_name varchar(255) NOT NULL DEFAULT '',
	matches_won integer DEFAULT 0,
	matches_drawn integer DEFAULT 0,
	matches_lost integer DEFAULT 0,
	games_won integer DEFAULT 0,
	games_lost integer DEFAULT 0,
	points_won integer DEFAULT 0,
	current_streak integer DEFAULT 0,
	elo integer DEFAULT 1000,
	UNIQUE(user_id, channel_id, team_id)
);
