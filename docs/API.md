# API

## Create a Leaderboard

**Path:** `/leaderboards`

**Method:** `POST`

**Headers:**

- `Content-Type: application/x-www-form-url-encoded`

**Request Body**:

```x-www-form-urlencoded
name=unique-leaderboard-name
```

## Register a Webhook on a Leaderboard

**Path:** `/leaderboards/{leaderboard_name}/webhooks`

**Method:** `POST`

**Headers:**

- `Content-Type: application/x-www-form-url-encoded`

**Request Body:**

```x-www-form-urlencoded
url=https://example.com/webhook-endpoint
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

- `Content-Type: application/x-www-form-url-encoded`

**Request Body:**

```x-www-form-urlencoded
username=player_username
```
## Record a Match Result

**Path:** `/leaderboards/{leaderboard_name}/matches`

**Method**: `POST`

**Headers:**

- `Content-Type: application/x-www-form-url-encoded`

**Request Body:**

```x-www-form-urlencoded
player1=username1&player2=username2&score=2-1
```

## Retrieve the Leaderboard

**Path:** `/leaderboards/{leaderboard_name}`

**Method:** `GET`

## Retrieve Player Stats

**Path:** `/leaderboards/{leaderboard_name}/players/{username}`

**Method:** `GET`

