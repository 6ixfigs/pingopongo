# Pongo CLI Commands

## Create a Leaderboard

**Command:**

```bash
pongo create-leaderboard <name>
```

**Description:** Creates a new leaderboard with the specified name.

## Add a Webhook to a Leaderboard

**Command:**

```bash
pongo register-webhook <leaderboard-name> <url>
```

**Description:** Registers a webhook for notifications on the specified leaderboard.

## List Webhooks for a Leaderboard

**Command:**

```bash
pongo list-webhooks <leaderboard-name>
```

**Description:** List all webhooks registered for the specified leaderboard.

## Delete all Webhooks from a Leaderboard

**Command:**

```bash
pongo delete-webhooks <leaderboard-name>
```

**Description:** Removes a webhook from the specified leaderboard.


## Create a Player on a Leaderboard

**Command:**

```bash
pongo create-player <leaderboard-name> <username>
```

**Description:** Creates a player on the specified leaderboard.

## Record a Match Result

**Command:**

```
pongo record <leaderboard-name> <player1> <player2> <score>
```

**Description:** Records a match result between two players.

## Retrieve a Leaderboard

**Command:**

```bash
pongo leaderboard <leaderboard-name>
```

**Description:** Retrieves the leaderboard.

## Retrieve Player Stats

**Command:**

```bash
pongo stats <leaderboard-name> <username>
```

**Description:** Retrieves stats for a specific player in the leaderboard.
