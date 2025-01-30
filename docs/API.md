# API

## Create a Leaderboard

**Path:** `/leaderboards`

**Method:** `POST`

**Request Body**:

```json
{
    "name": "unique-leaderboard-name"
}
```

## Add a Webhook to a Leaderboard

**Path:** `/leaderboards/{leaderboard_name}/webhooks`

**Method:** `POST`

**Request Body:**

```json
{
    "url": "https://example.com/webhook-endpoint"
}
```

## List Webhooks for a Leaderboard

**Path:** `/leaderboards/{leaderboard_name}/webhooks`

**Method:** `GET`

## Delete all Webhooks from a Leaderboard

**Path:** `/leaderboards/{leaderboard_name}/webhooks`

**Method:** `DELETE`

## Create a Player on a Leaderboard

**Path:** `/leaderboards/{leaderboard_name}/players`

**Method:** `POST`

**Headers:**

- `Content-Type: application/x-www-form-urlencoded`

**Request Body:**

```json
{
    "username": "player_username"
}
```
## Record a Match Result

**Path:** `/leaderboards/{leaderboard_name}/matches`

**Method**: `POST`

**Headers:**

- `Content-Type: application/x-www-form-urlencoded`

**Request Body:**

```json
{
    "player1": "username1"
    "player2": "username2"
    "scores": ["11-9", "14-12", "3-11"]
}
```

## Retrieve the Leaderboard

**Path:** `/leaderboards/{leaderboard_name}`

**Method:** `GET`

**Query Parameters:**

- `limit`: Optional - Number of top players to return. If 0, return all players.

## Retrieve Player Stats

**Path:** `/leaderboards/{leaderboard_name}/players/{username}`

**Method:** `GET`

