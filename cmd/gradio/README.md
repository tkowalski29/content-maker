# gradio CLI

`gradio` automates image generation through a custom Gradio environment. The tool expects a JSON file with a list of requests and saves image files along with metadata. Internally uses `internal/gradio/runner.go`.

## Requirements

- Running Gradio backend compatible with the API used by `internal/gradio`.
- Go 1.21+ and internet access (for downloading images).

## Running

```bash
# one-time execution
go run ./cmd/gradio -input requests.json -gradio_url https://your-instance.gradio.live

# or after building

go build -o gradio ./cmd/gradio
./gradio -input requests.json -gradio_url https://your-instance.gradio.live
```

## Input Format

JSON file should have structure matching `internal/gradio/InputData`:

```json
{
  "folder_name": "output/images",
  "items": [
    {
      "id": "HERO",
      "alt": "Hero image",
      "prompt": "A futuristic city at sunrise",
      "style": "photorealistic",
      "aspect_ratio": "16:9",
      "position_in_article": 120,
      "platform": "web"
    }
  ]
}
```

Each `items` element becomes a separate image saved in the `folder_name` directory.

## Operation Flow

1. Creates target directory (if it doesn't exist).
2. For each request:
   - Calculates resolution from `aspect_ratio` field (`16:9`, `1200x627`, etc.).
   - Calls `gradio_api/call/generate_image` endpoint and waits for event stream.
   - Saves image (`<ID>.png`) and metadata (`<ID>.json`).
3. Logs progress to stdout/stderr. Exit code `0` means all requests succeeded, `1` if any generation failed.

## Output Files

- `folder_name/<ID>.png` – generated image.
- `folder_name/<ID>.json` – metadata (`request`, used seed, generation time, Gradio payload).

## Troubleshooting

- Full stream URLs and SSE events are logged – this helps debug backend errors.
- If you see an HTTP status error message, check the `-gradio_url` parameter and instance availability.

## Tests

Simple JSON structure test is in `cmd/gradio/main_test.go` and is based on types from `internal/gradio`.
