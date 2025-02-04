# Pongo CLI
This is a command line CLI app for tracking ping pong matches.

## Supported commands
- `pongo record <leaderboard-name> <player1> <player2> <score>`
  - Aliases: *r*, *rec*
  - Records a match between two players inside the group specified with <leaderboard-name> 
- `pongo create-leaderboard <leaderboard-name>`
  - Aliases: *cl*
  - Creates a group (leaderboard) witht the specified name
  
- `pongo create-player <leaderboard-name> <username>`
    - Aliases: *cp*, *new-player*
    - Creates a player inside the specified group (leaderboard) with the given username 

- `pongo register-webhook <leaderboard-name> <url>`
  - Aliases: *rw*, *reg*
  - Registers a webhook url where the server will send responses for the user to see

- `pongo list-webhooks <leaderboard-name>`
  - Aliases: *lw*, *list*
  - Lists all webhook urls registered to the specified leaderboard

- `pongo delete-webhooks <leaderboard-name>`
  - Aliases: *dw*, 
  - Deletes all webhook urls registered to the specified leaderboard
    - User will be prompted an 'Are you sure you want to delete?' question before deletion

- `pongo leaderboard <leaderboard-name>`
  - Aliases: *l*
  - Displayes the rankings inside the specified group (leaderboard)

- `pongo stats <leaderboard-name> <username>`
  - Aliases: *s*
  - Displays players' stats inside the specified group (leaderboard)

- `pongo example`
  - Aliases: *e*
  - Shows an example of how to use Pongo CLI


  