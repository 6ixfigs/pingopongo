# PingyPongy

PingyPongy is an open-source application for tracking table tennis scores, integrated with any platform (like Slack, Discord, MS Teams).

## Core Features

### Match Recording

**Description:** Record a match result by specifying the players and the final game outcome.

**Command:** `pongo record <leaderboard-name> <player1> <player2> <player1-score> - <player2-score>`

**Example:**

```
pongo record marc vux 3-1
```

**Response:**

```
Match recorded successfully:
marc vs vux
🎉 Winner: marc (3-1 in sets)
```

### Leaderboard

**Description:** Display the current leaderboard.

**Command:** `pongo leaderboard`

**Response:**

```
🏓 Current Leaderboard:
Rank | Player   | Won | Lost | Drawn | Played | Win Ratio
-------------------------------------------------
1    | John     | 5   | 1    | 0     | 6      | 83.33%
2    | Jane     | 4   | 2    | 0     | 6      | 66.67%
3    | Bob      | 3   | 3    | 0     | 6      | 50.00%
4    | Alice    | 1   | 5    | 0     | 6      | 16.67%
```

### Player Stats

**Description:** View individual player stats like win/loss ratio, matches won/lost, etc.

**Command:** `pongo stats <player>`

**Example:** `pongo stats marc`

**Response:**

```
Stats for marc:
- Matches Played: 6
- Matches Won: 5
- Matches Lost: 1
- Games Won: 11
- Games Lost: 5
- Win Ratio: 83.33%
- Current Streak: 4 Wins
```
