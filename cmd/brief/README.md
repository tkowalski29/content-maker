# brief CLI

`brief` is a command-line tool for content analysis and working with SERP data. Each command delegates logic to modules in the `internal/brief` directory, allowing it to be used as a library as well.

## Running

```bash
# from repository directory
go run ./cmd/brief <subcommand> [options]
```

You can also build the binary:

```bash
go build -o brief ./cmd/brief
./brief <subcommand> [options]
```

## Subcommands

| Subcommand | Description | Key flags |
|------------|-------------|-----------|
| `analyze`  | Aggregates URL audits into a content gap report. | `-audits` (path to `all_audits.json`), `-output` (optional save path) |
| `fetch`    | Fetches and parses HTML content using selected method. | `-url`, `-method` (`jina`, `netlify`, `direct`), `-output` |
| `serp`     | Retrieves search results from SearXNG instance. | `-query`, `-lang`, `-country`, `-limit`, `-searxng-url` |
| `suggest`  | Fetches keyword suggestions. | `-query`, `-provider` (`google`, `bing`, `dgd`), `-lang`, `-country` |

### analyze

```bash
go run ./cmd/brief analyze -audits data/all_audits.json -output analysis.json
```

Input file should be an array of audit objects (see `internal/brief/analyze/types.go`).

### fetch

```bash
go run ./cmd/brief fetch \
  -url https://example.com \
  -method jina \
  -output page.json
```

- `jina` – uses [Jina Reader](https://r.jina.ai/)
- `netlify` – uses custom Netlify function (see variable in `internal/brief/fetch/service.go`)
- `direct` – direct HTTP request with simple HTML parsing.

### serp

```bash
go run ./cmd/brief serp \
  -searxng-url "https://" \
  -query "best beaches in spain" \
  -lang en \
  -country US \
  -limit 10
```

### suggest

```bash
go run ./cmd/brief suggest \
  -query "beaches" \
  -provider google \
  -lang en \
  -country US
```

Output contains a list of suggestions and a timestamp generated in UTC.

## Output and Errors

- Upon successful execution, each command exits with code `0` and outputs JSON to stdout (or to a file if `-output` is provided).
- Argument validation and operation errors are written to stderr and exit with code `1`.

## Dependencies

The command uses only the Go standard library. You can find types and domain functions in the `internal/brief` directory.
