CREATE TABLE players (
	id integer GENERATED ALWAYS AS IDENTITY,
	user_id varchar(255) NOT NULL,
	channel_id varchar(255) NOT NULL,
	matches_won integer,
	matches_drawn integer,
	matches_lost integer,
	sets_won integer,
	sets_lost integer,
	current_streak integer,
	UNIQUE(user_id, channel_id)
);
