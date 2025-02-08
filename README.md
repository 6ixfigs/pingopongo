# Pingo-Pongo 🏓

Pingo-Pongo is a tool for tracking ping pong scores, player stats and leaderboards. It also supports registering webhooks on leaderboards, enabling members to receive notifications about match results and leaderboard changes.

- **API**: `pongo` - Manage scores, leaderboards, players, and webhooks programatically.
- **CLI**: `pingo` - Interact with the API from the command line.

## Installation

### CLI (`pingo`)

1. Download the latest binary from the [Releases](https://github.com/6ixfigs/pingopongo/releases) page.
2. Make the binary executable:

```bash
chmod +x pingo
```

3. Move the binary to a directory in your `PATH` (Optional):

```bash
mv pingo /usr/local/bin/
```

4. Verify the installation:

```bash
pingo version
```

### API (`pongo`)

1. Clone the repository:

```bash
git clone https://github.com/6ixfigs/pingopongo.git
cd pingopongo
```

2. Configure `.env` based on `.env.example`.

3. Start the services using `docker compose` or `make`:

```bash
docker compose up -d --build # or: make up
```

4. Apply database migrations:

```bash
docker compose --profile tools run --rm migrate up # or: make migrate-up
```

## CLI Usage

The `pingo` CLI allows you to interact with the Pongo API from the command line.

```bash
$ pingo
CLI for interacting with the Pongo server

Usage:
  pingo [command]

Available Commands:
  help        Help about any command
  leaderboard Create or retrieve leaderboards
  player      Create a player or retrieve stats
  record      Record a match between two players
  version     Print Pingo version number
  webhooks    Manage webhooks

Flags:
  -h, --help   help for pingo

Use "pingo [command] --help" for more information about a command.
```

## API Usage

The `pongo` API provides endpoints for managing scores, players, leaderboards and webhooks. Refer to the [API Documentation](docs/API.md) for details.

## Contributing

We welcome contributions! Please read the [Contribution Guidelines](CONTRIBUTING.md) to get started.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE.md) for details.
