# Ravi Reader

Ravi Reader is a fast, lightweight Go web application that aggregates horse racing ("ravi") data from the Finnish betting agency [Veikkaus](https://www.veikkaus.fi/).

To ensure lightning-fast UI responses and avoid hitting Veikkaus API rate limits, the application fetches data in the background every 5 minutes and caches it locally using **DuckDB**.

## Prerequisites

Because DuckDB relies on high-performance C++ bindings, you must have CGO enabled and a C compiler installed on your system.

- **Go 1.26+**
- **C Compiler** (e.g., `gcc` or `clang`). On Ubuntu/Debian, install with:
  ```bash
  sudo apt update && sudo apt install build-essential
  ```

## Getting Started

1. **Clone the repository**
   ```bash
   git clone https://github.com/mikcompany/ravireader.git
   cd ravireader
   ```

2. **Run the Server**
   To start the web server and the background data updater, simply run:
   ```bash
   go run main.go
   ```

3. **View the Application**
   - Web Interface: [http://localhost:8080](http://localhost:8080)
   - JSON Data: [http://localhost:8080/json](http://localhost:8080/json)

## Database & DuckDB Quack Protocol

The data is stored locally in a highly structured `data.duckdb` file (with tables for `cards`, `races`, `results`, and `runners`).

If you want to connect to the database concurrently (for example, to run `dbt` transformations or to query the data from the DuckDB CLI while the server is running), you can start the application with the **Quack protocol server** enabled:

```bash
go run main.go --serve-quack
```

This exposes the database on `localhost:9494`. You can connect to it securely using the DuckDB CLI:
```sql
ATTACH 'quack:localhost:9494' AS ravireader_db (TOKEN 'ravitoken');
SELECT * FROM quack_query('quack:localhost:9494', 'SELECT * FROM cards', token = 'ravitoken');
```

Alternatively, you can connect the Go application to an *external* Quack server by setting the environment variable:
```bash
DUCKDB_QUACK_URL="quack:hostname:9494" go run main.go
```

## Running Tests

To run the unit tests (which safely mock the Veikkaus API using an in-memory database), run:
```bash
go test ./...
```

## Contributing & Releasing

This project strictly enforces **[Conventional Commits](https://www.conventionalcommits.org/)** (e.g., `feat:`, `fix:`, `chore:`). If your commit message does not match this format, the local git hook will block it.

To create a new semantic version release:
1. Ensure your working directory is clean.
2. Run the automated release script:
   ```bash
   ./scripts/release.sh
   ```
3. The script will automatically calculate the next version number based on your commits and generate a `CHANGELOG.md` file.
4. Review the changelog, then commit and tag the release:
   ```bash
   git commit -am "chore(release): vX.X.X"
   git tag vX.X.X
   git push --follow-tags
   ```
