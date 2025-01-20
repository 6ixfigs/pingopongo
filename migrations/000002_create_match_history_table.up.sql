CREATE TABLE match_history (
	id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	player1_id integer NOT NULL,
	player2_id integer NOT NULL,
	player1_games_won integer NOT NULL,
	player2_games_won integer NOT NULL,
	CONSTRAINT fk_player1 FOREIGN KEY (player1_id) REFERENCES players(id),
	CONSTRAINT fk_player2 FOREIGN KEY (player2_id) REFERENCES players(id)
);
