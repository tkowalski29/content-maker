# colab CLI

`colab` runs an automated scenario that controls Google Colab notebooks using Playwright. It uses user configuration from an `env.json` file (compatible with `automation/colab-go`).

## Requirements

- Playwright and Chrome/Chromium available locally.
- `env.json` file with user configuration (format identical to `automation/colab-go/env.json.example`). Path to this file is passed as a parameter.
- Go 1.21+ and `github.com/playwright-community/playwright-go` dependency (added to `go.mod`).

## `env.json` Format

```json
{
  "webhook_url": "https://example.com/hooks/gradio",
  "users": [
    {
      "id": "test1",
      "email": "user@example.com",
      "password": "secret",
      "colab_url": "https://colab.research.google.com/drive/..."
    }
  ]
}
```

`id` corresponds to the `-user` parameter passed to the CLI.

## Running

```bash
# one-time execution
go run ./cmd/colab -env <path> -user <user_id>

# or after building

go build -o colab ./cmd/colab
./colab -env <path> -user <user_id>

# examples
./colab -env ./env.json -user test1
./colab -env /path/to/config/env.json -user test1
```

On first run, manual Google login is required in the opened browser. The session will be saved in `chrome-user-data/<user_id>`.

## Step-by-Step Operation

1. Kills existing Chrome processes associated with the `chrome-user-data` directory.
2. Starts Playwright in visible mode, using the `chrome-user-data/<user_id>` profile.
3. Opens the specified Colab notebook.
4. Allows manual login (max 5 minutes). Subsequent runs use the saved session.
5. Changes runtime type to T4 GPU, saves settings, and runs `Run all`.
6. Listens for logs/responses from `gradio.live`, then sends the found URL to the webhook (if configured).
7. Starts keep-alive mechanism (scroll + events) and leaves the browser window open for manual closure.

Screenshots, HTML, and logs are saved in `debug/<date>/<time>/`.

## Output and Error Codes

- Success → code `0`.
- Missing required data (`env.json`, user configuration, Playwright errors) → code `1`, details on stderr.

## Additional Information

Full logic (including keep-alive, webhook handling, and session management) is in `internal/colab/runner.go`. The file was migrated 1:1 from `automation/colab-go/main.go`, preserving original functionality.
