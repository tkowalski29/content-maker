package gradio

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ImageRequest struct {
	ID                string `json:"id"`
	Alt               string `json:"alt"`
	Prompt            string `json:"prompt"`
	Style             string `json:"style"`
	AspectRatio       string `json:"aspect_ratio"`
	PositionInArticle int    `json:"position_in_article,omitempty"`
	Platform          string `json:"platform,omitempty"`
}

type InputData struct {
	FolderName string         `json:"folder_name"`
	Items      []ImageRequest `json:"items"`
}

type GradioRequest struct {
	Data []interface{} `json:"data"`
}

type GradioResponse []interface{}

type GradioPayload struct {
	Prompt    string `json:"prompt"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Seed      int    `json:"seed"`
	Steps     int    `json:"steps"`
	Sampler   string `json:"sampler"`
	Scheduler string `json:"scheduler"`
}

type StreamImageResult struct {
	Image []byte
	Seed  *int64
}

type ImageOutputMetadata struct {
	Request       ImageRequest  `json:"request"`
	GradioPayload GradioPayload `json:"gradio_payload"`
	Seed          *int64        `json:"seed,omitempty"`
	DurationMs    int64         `json:"duration_ms"`
	EventID       string        `json:"event_id,omitempty"`
	GeneratedAt   string        `json:"generated_at"`
}

const (
	requestTimeout  = 25 * time.Minute
	streamTimeout   = 25 * time.Minute
	downloadTimeout = 2 * time.Minute
)

// Runner orchestrates image generation requests against a Gradio backend.
type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
	Client *http.Client
}

// Result captures outcome for a single request.
type Result struct {
	Request ImageRequest
	Err     error
}

// NewRunner creates a Runner with default timeouts.
func NewRunner(stdout, stderr io.Writer) *Runner {
	return &Runner{
		Stdout: stdout,
		Stderr: stderr,
		Client: &http.Client{Timeout: requestTimeout},
	}
}

// RunFile loads configuration from disk and executes generation for each item.
func (r *Runner) RunFile(path string, gradioURL string) ([]Result, error) {
	if path == "" {
		return nil, fmt.Errorf("input path is required")
	}

	if gradioURL == "" {
		return nil, fmt.Errorf("gradio URL is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading input file: %w", err)
	}

	var input InputData
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if err := os.MkdirAll(input.FolderName, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	results := r.Generate(input, gradioURL)

	var firstErr error
	for _, res := range results {
		if res.Err != nil && firstErr == nil {
			firstErr = res.Err
		}
	}

	return results, firstErr
}

// Generate processes each image request and returns individual outcomes.
func (r *Runner) Generate(input InputData, gradioURL string) []Result {
	results := make([]Result, 0, len(input.Items))

	for _, req := range input.Items {
		err := r.handleRequest(req, gradioURL, input.FolderName)
		if err != nil {
			fmt.Fprintf(r.Stderr, "Error processing %s: %v\n", req.ID, err)
			results = append(results, Result{Request: req, Err: err})
			continue
		}
		fmt.Fprintf(r.Stdout, "✓ Generated image: %s\n", req.ID)
		results = append(results, Result{Request: req, Err: nil})
	}

	return results
}

func (r *Runner) handleRequest(req ImageRequest, gradioURL, outputDir string) error {
	width, height := parseAspectRatioToDimensions(req.AspectRatio)
	fmt.Fprintf(r.Stdout, "  Using dimensions: %dx%d (aspect_ratio: %s)\n", width, height, req.AspectRatio)

	payload := GradioPayload{
		Prompt:    req.Prompt,
		Width:     width,
		Height:    height,
		Seed:      0,
		Steps:     20,
		Sampler:   "euler",
		Scheduler: "simple",
	}

	start := time.Now()
	streamResult, eventID, err := r.callGradioAPI(gradioURL, payload)
	if err != nil {
		return fmt.Errorf("calling Gradio API: %w", err)
	}
	generationDuration := time.Since(start)
	generatedAt := time.Now().UTC()

	filename := fmt.Sprintf("%s.%s", req.ID, "png")
	outputPath := filepath.Join(outputDir, filename)

	if err := os.WriteFile(outputPath, streamResult.Image, 0644); err != nil {
		return fmt.Errorf("saving image: %w", err)
	}

	metadata := ImageOutputMetadata{
		Request:       req,
		GradioPayload: payload,
		Seed:          streamResult.Seed,
		DurationMs:    generationDuration.Milliseconds(),
		EventID:       eventID,
		GeneratedAt:   generatedAt.Format(time.RFC3339),
	}

	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling metadata: %w", err)
	}

	metadataFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".json"
	metadataPath := filepath.Join(outputDir, metadataFilename)
	if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
		return fmt.Errorf("saving metadata: %w", err)
	}

	return nil
}

func (r *Runner) callGradioAPI(gradioURL string, payload GradioPayload) (*StreamImageResult, string, error) {
	apiURL := strings.TrimSuffix(gradioURL, "/") + "/gradio_api/call/generate_image"

	requestBody := GradioRequest{
		Data: []interface{}{
			payload.Prompt,
			payload.Width,
			payload.Height,
			payload.Seed,
			payload.Steps,
			payload.Sampler,
			payload.Scheduler,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, "", err
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: requestTimeout}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var eventResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&eventResp); err != nil {
		return nil, "", err
	}

	eventID, ok := eventResp["event_id"].(string)
	if !ok {
		return nil, "", fmt.Errorf("missing event_id in response")
	}

	fmt.Fprintf(r.Stdout, "  Event ID: %s\n", eventID)

	result, err := r.streamGradioResult(gradioURL, eventID)
	if err != nil {
		return nil, "", err
	}
	if result.Seed != nil {
		fmt.Fprintf(r.Stdout, "  Used seed: %d\n", *result.Seed)
	}

	return result, eventID, nil
}

func (r *Runner) streamGradioResult(gradioURL, eventID string) (*StreamImageResult, error) {
	streamURL := fmt.Sprintf("%s/gradio_api/call/generate_image/%s", strings.TrimSuffix(gradioURL, "/"), eventID)

	fmt.Fprintf(r.Stdout, "  Streaming from: %s\n", streamURL)

	client := r.Client
	timeout := streamTimeout
	if client == nil {
		client = &http.Client{}
	}
	originalTimeout := client.Timeout
	client.Timeout = timeout
	defer func() { client.Timeout = originalTimeout }()

	req, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stream api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	var (
		currentEvent string
		dataLines    []string
		streamDone   bool
	)

	processPayload := func(eventName, payload string) (*StreamImageResult, error) {
		payload = strings.TrimSpace(payload)
		if payload == "" {
			return nil, nil
		}
		if payload == "[DONE]" {
			return nil, io.EOF
		}

		var raw interface{}
		if err := json.Unmarshal([]byte(payload), &raw); err != nil {
			if eventName == "" {
				eventName = "message"
			}
			fmt.Fprintf(r.Stdout, "  Ignoring %s event (unparseable payload): %v\n", eventName, err)
			return nil, nil
		}

		handleDataArray := func(arr []interface{}) (*StreamImageResult, error) {
			if len(arr) == 0 {
				return nil, nil
			}
			return extractImageFromResponse(gradioURL, GradioResponse(arr))
		}

		switch v := raw.(type) {
		case map[string]interface{}:
			if errMsg, ok := v["error"].(string); ok && errMsg != "" {
				return nil, fmt.Errorf("gradio stream error (%s): %s", eventName, errMsg)
			}
			if rawData, ok := v["data"]; ok {
				if arr, ok := rawData.([]interface{}); ok {
					return handleDataArray(arr)
				}
			}
			if outputs, ok := v["output"].([]interface{}); ok {
				return handleDataArray(outputs)
			}
			if body, ok := v["body"].([]interface{}); ok {
				return handleDataArray(body)
			}
		case []interface{}:
			return handleDataArray(v)
		case string:
			if eventName == "" {
				eventName = "message"
			}
			if strings.HasPrefix(v, "data:image") || strings.HasPrefix(v, "http") || strings.HasPrefix(v, "file=") {
				imageData, err := extractImageFromValue(gradioURL, v)
				if err != nil {
					return nil, err
				}
				return &StreamImageResult{Image: imageData}, nil
			}
			return nil, nil
		case nil:
			return nil, nil
		default:
			if eventName == "" {
				eventName = "message"
			}
			fmt.Fprintf(r.Stdout, "  Ignoring %s event (unsupported payload type %T)\n", eventName, v)
		}

		return nil, nil
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")

		switch {
		case strings.HasPrefix(line, "event:"):
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			dataLine := strings.TrimPrefix(line, "data:")
			if len(dataLine) > 0 && dataLine[0] == ' ' {
				dataLine = dataLine[1:]
			}
			dataLines = append(dataLines, dataLine)
		case line == "":
			if len(dataLines) > 0 {
				eventName := currentEvent
				if eventName == "" {
					eventName = "message"
				}
				payload := strings.Join(dataLines, "\n")
				dataLines = nil
				if result, processErr := processPayload(eventName, payload); processErr != nil {
					if errors.Is(processErr, io.EOF) {
						streamDone = true
					} else {
						return nil, processErr
					}
				} else if result != nil {
					return result, nil
				}
			}
			currentEvent = ""
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	if len(dataLines) > 0 {
		payload := strings.Join(dataLines, "\n")
		eventName := currentEvent
		if eventName == "" {
			eventName = "message"
		}
		if result, err := processPayload(eventName, payload); err != nil {
			if errors.Is(err, io.EOF) {
				streamDone = true
			} else {
				return nil, err
			}
		} else if result != nil {
			return result, nil
		}
	}

	if streamDone {
		return nil, fmt.Errorf("stream ended before delivering image data")
	}

	return nil, fmt.Errorf("no image data received from stream")
}

func extractImageFromResponse(baseURL string, resp GradioResponse) (*StreamImageResult, error) {
	if len(resp) < 1 {
		return nil, fmt.Errorf("empty response")
	}

	imageData, err := extractImageFromValue(baseURL, resp[0])
	if err != nil {
		return nil, err
	}

	var seedPtr *int64
	if len(resp) > 1 {
		if parsedSeed := extractSeedValue(resp[1]); parsedSeed != nil {
			seedPtr = parsedSeed
		}
	}

	return &StreamImageResult{Image: imageData, Seed: seedPtr}, nil
}

func extractImageFromValue(baseURL string, value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		if urlVal, ok := v["url"].(string); ok {
			return fetchImageFromString(baseURL, urlVal)
		}
		if pathVal, ok := v["path"].(string); ok {
			return fetchImageFromString(baseURL, pathVal)
		}
		if dataVal, ok := v["data"].(string); ok {
			return fetchImageFromString(baseURL, dataVal)
		}
		return nil, fmt.Errorf("could not find image data in response map")
	case string:
		return fetchImageFromString(baseURL, v)
	default:
		return nil, fmt.Errorf("unsupported image payload type %T", v)
	}
}

func fetchImageFromString(baseURL, value string) ([]byte, error) {
	if strings.HasPrefix(value, "file=") {
		value = strings.TrimPrefix(value, "file=")
	}

	switch {
	case strings.HasPrefix(value, "data:image"):
		parts := strings.SplitN(value, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL")
		}
		return base64.StdEncoding.DecodeString(parts[1])
	case strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://"):
		return downloadImageFromURL(value)
	default:
		trimmedBase := strings.TrimSuffix(baseURL, "/")
		var absolute string
		if strings.HasPrefix(value, "/") {
			absolute = trimmedBase + value
		} else {
			absolute = trimmedBase + "/" + value
		}
		if strings.HasPrefix(absolute, "http://") || strings.HasPrefix(absolute, "https://") {
			return downloadImageFromURL(absolute)
		}
		return nil, fmt.Errorf("unsupported image reference %q", value)
	}
}

func downloadImageFromURL(url string) ([]byte, error) {
	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func extractSeedValue(value interface{}) *int64 {
	switch v := value.(type) {
	case float64:
		seed := int64(v)
		return &seed
	case int:
		seed := int64(v)
		return &seed
	case int32:
		seed := int64(v)
		return &seed
	case int64:
		seed := v
		return &seed
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return &i
		}
	case string:
		if s, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return &s
		}
	case map[string]interface{}:
		if innerSeed, ok := v["seed"]; ok {
			return extractSeedValue(innerSeed)
		}
	}

	return nil
}

func parseAspectRatioToDimensions(aspectRatio string) (int, int) {
	if strings.Contains(aspectRatio, "x") {
		parts := strings.Split(aspectRatio, "x")
		if len(parts) == 2 {
			var width, height int
			fmt.Sscanf(parts[0], "%d", &width)
			fmt.Sscanf(parts[1], "%d", &height)
			return width, height
		}
	}

	if strings.Contains(aspectRatio, ":") {
		parts := strings.Split(aspectRatio, ":")
		if len(parts) == 2 {
			var w, h int
			fmt.Sscanf(parts[0], "%d", &w)
			fmt.Sscanf(parts[1], "%d", &h)

			ratioMap := map[string][2]int{
				"16:9": {1920, 1080},
				"9:16": {1080, 1920},
				"4:3":  {1600, 1200},
				"3:4":  {1200, 1600},
				"1:1":  {1024, 1024},
				"21:9": {2560, 1080},
				"2:3":  {1200, 1800},
				"3:2":  {1800, 1200},
			}

			if dims, ok := ratioMap[aspectRatio]; ok {
				return dims[0], dims[1]
			}

			scale := 1024.0
			return int(float64(w) * scale / float64(h)), int(float64(h) * scale / float64(h))
		}
	}

	return 1024, 1024
}
