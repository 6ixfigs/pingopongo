# Contributing to Pingo-Pongo

Thank you for your interest in contributing to Pingo-Pongo! Whether you’re fixing a bug, adding a feature, or improving documentation, your contributions are welcome.

## Getting Started

### Prerequisites

- [Go](https://golang.org/) (version 1.23.4 or higher).
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/).

### Setup  

1. Clone the repository.

```bash
git clone https://github.com/6ixfigs/pingopongo.git
```

2. Navigate to the repository directory on your file system.

3. Setup `.env` based on [.env.template](.env.template).

4. Build and start the services using `docker compose`.

```bash
docker compose up -d --build
```

5. Apply database migrations.

```bash
docker compose --profile tools run --rm migrate up
```

## Useful Make Commands

Here are some useful `make` commands for more efficient development:

| Command               | Description                                                |
|-----------------------|------------------------------------------------------------|
| `make up`             | Build and run `app` and `db` containers in the background. |
| `make down`           | Stop and remove all containers.                            |
| `make watch`          | Build and run `app` and `db` containers in the foreground. |
| `make migrate-up`     | Apply database migrations.                                 |
| `make migrate-down`   | Undo database migrations.                                  |
| `make migrate-create` | Create a new database migration file.                      |
| `make migrate-force`  | Force a database migration version.                        |
| `make db-shell`       | Connect to the database.                                   |
| `make db-exec`        | Execute an SQL query on the database.                      |

## Making Changes

1. Create a new branch for your changes:

```bash
git checkout -b feature/your-feature-name
```

2. Make your changes and test them thoroughly.

3. Commit your changes with a clear and descriptive message:

```bash
git commit -m "Add: new feature to view match history"
```

4. Push your changes to the remote repository.

```
git push origin feature/your-feature-name
```

## Submitting a Pull Request

1. Open a pull request (PR) on GitHub.

2. Provide a clear title and description for your PR.
- Explain the problem you're solving or the feature you're adding.
- Include any relevant issue numbers.

3. Wait for feedback. Your PR may be reviewed and dicussed before being merged.

## Code Style

- Follow the existing code style and formatting.
- Use `gofmt` to format your code.
- Write clear and concise commit messages.

## Reporting Issues

If you find a bug or have a suggestion, please open an issue using one of the templates:
- [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md)
- [Feature Request](.github/ISSUE_TEMPLATE/feature_request.md)

## Maintenance

This project is maintained on a best-effort basis. If you encounter an issue or have a feature request, feel free to open a PR!.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE.md).
