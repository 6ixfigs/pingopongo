# PingyPongy

PingyPongy is an application for tracking Ping Pong scores, integrated with Slack.

## Core Features

### Match Recording

**Description:** Record a match result by specifying the players and the score in each set.

**Command:** `/pingypongy record <player1> <player2> <game1> [games...]`

**Example:**

```
/pingypongy record @marc @vukota 11-7 5-11 11-8
```

**Response:**

```
Match recorded successfully:
@marc vs @vukota
- Game 1: 11-7
- Game 2: 5-11
- Game 3: 11-8
🎉 Winner: @marc (2-1 in sets)
```

### Leaderboard

**Description:** Display the current leaderboard.

**Command:** `/pingypongy leaderboard`

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

**Command:** `/pingypongy stats <player>`

**Example:** `/pingypongy stats @marc`

**Response:**

```
Stats for @marc:
- Matches Played: 6
- Matches Won: 5
- Matches Lost: 1
- Games Won: 11
- Games Lost: 5
- Win Ratio: 83.33%
- Current Streak: 4 Wins
```
