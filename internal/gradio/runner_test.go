package gradio

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAspectRatioToDimensions(t *testing.T) {
	tests := []struct {
		name           string
		aspectRatio    string
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "16:9 ratio",
			aspectRatio:    "16:9",
			expectedWidth:  1920,
			expectedHeight: 1080,
		},
		{
			name:           "9:16 ratio (vertical)",
			aspectRatio:    "9:16",
			expectedWidth:  1080,
			expectedHeight: 1920,
		},
		{
			name:           "4:3 ratio",
			aspectRatio:    "4:3",
			expectedWidth:  1600,
			expectedHeight: 1200,
		},
		{
			name:           "1:1 square",
			aspectRatio:    "1:1",
			expectedWidth:  1024,
			expectedHeight: 1024,
		},
		{
			name:           "21:9 ultrawide",
			aspectRatio:    "21:9",
			expectedWidth:  2560,
			expectedHeight: 1080,
		},
		{
			name:           "explicit dimensions with x",
			aspectRatio:    "1920x1080",
			expectedWidth:  1920,
			expectedHeight: 1080,
		},
		{
			name:           "explicit dimensions 400x400",
			aspectRatio:    "400x400",
			expectedWidth:  400,
			expectedHeight: 400,
		},
		{
			name:           "explicit dimensions 1200x627",
			aspectRatio:    "1200x627",
			expectedWidth:  1200,
			expectedHeight: 627,
		},
		{
			name:           "invalid format defaults to square",
			aspectRatio:    "invalid",
			expectedWidth:  1024,
			expectedHeight: 1024,
		},
		{
			name:           "empty string defaults to square",
			aspectRatio:    "",
			expectedWidth:  1024,
			expectedHeight: 1024,
		},
		{
			name:           "2:3 ratio",
			aspectRatio:    "2:3",
			expectedWidth:  1200,
			expectedHeight: 1800,
		},
		{
			name:           "3:2 ratio",
			aspectRatio:    "3:2",
			expectedWidth:  1800,
			expectedHeight: 1200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := parseAspectRatioToDimensions(tt.aspectRatio)

			if width != tt.expectedWidth {
				t.Errorf("width = %d, want %d", width, tt.expectedWidth)
			}
			if height != tt.expectedHeight {
				t.Errorf("height = %d, want %d", height, tt.expectedHeight)
			}
		})
	}
}

func TestInputDataStructure(t *testing.T) {
	// Test that InputData structure works as expected
	input := InputData{
		FolderName: "test/output",
		Items: []ImageRequest{
			{
				ID:                "IMAGE_1",
				Alt:               "Test alt",
				Prompt:            "Test prompt",
				Style:             "photorealistic",
				AspectRatio:       "16:9",
				PositionInArticle: 100,
			},
		},
	}

	if input.FolderName != "test/output" {
		t.Errorf("FolderName = %q, want %q", input.FolderName, "test/output")
	}

	if len(input.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(input.Items))
	}

	item := input.Items[0]
	if item.ID != "IMAGE_1" {
		t.Errorf("ID = %q, want %q", item.ID, "IMAGE_1")
	}
	if item.Style != "photorealistic" {
		t.Errorf("Style = %q, want %q", item.Style, "photorealistic")
	}
}

func TestGradioPayloadStructure(t *testing.T) {
	// Test GradioPayload structure
	payload := GradioPayload{
		Prompt:    "Test prompt",
		Width:     1920,
		Height:    1080,
		Seed:      0,
		Steps:     20,
		Sampler:   "euler",
		Scheduler: "simple",
	}

	if payload.Width != 1920 {
		t.Errorf("Width = %d, want %d", payload.Width, 1920)
	}
	if payload.Height != 1080 {
		t.Errorf("Height = %d, want %d", payload.Height, 1080)
	}
	if payload.Steps != 20 {
		t.Errorf("Steps = %d, want %d", payload.Steps, 20)
	}
	if payload.Sampler != "euler" {
		t.Errorf("Sampler = %q, want %q", payload.Sampler, "euler")
	}
}

func TestExtractSeedValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected *int64
	}{
		{
			name:     "float64 seed",
			input:    float64(12345),
			expected: func() *int64 { v := int64(12345); return &v }(),
		},
		{
			name:     "int seed",
			input:    int(99999),
			expected: func() *int64 { v := int64(99999); return &v }(),
		},
		{
			name:     "int64 seed",
			input:    int64(88888),
			expected: func() *int64 { v := int64(88888); return &v }(),
		},
		{
			name:     "int32 seed",
			input:    int32(77777),
			expected: func() *int64 { v := int64(77777); return &v }(),
		},
		{
			name:     "string seed",
			input:    "54321",
			expected: func() *int64 { v := int64(54321); return &v }(),
		},
		{
			name:     "string with spaces",
			input:    "  12345  ",
			expected: func() *int64 { v := int64(12345); return &v }(),
		},
		{
			name:     "json.Number",
			input:    json.Number("99999"),
			expected: func() *int64 { v := int64(99999); return &v }(),
		},
		{
			name:     "map with seed key",
			input:    map[string]interface{}{"seed": int64(88888)},
			expected: func() *int64 { v := int64(88888); return &v }(),
		},
		{
			name:     "invalid string",
			input:    "not-a-number",
			expected: nil,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: nil,
		},
		{
			name:     "bool input",
			input:    true,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSeedValue(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", *result)
				}
				return
			}

			if result == nil {
				t.Error("Expected non-nil result, got nil")
				return
			}

			if *result != *tt.expected {
				t.Errorf("result = %d, want %d", *result, *tt.expected)
			}
		})
	}
}

func TestNewRunner(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runner := NewRunner(stdout, stderr)

	if runner == nil {
		t.Fatal("NewRunner() returned nil")
	}

	if runner.Stdout != stdout {
		t.Error("Stdout not set correctly")
	}

	if runner.Stderr != stderr {
		t.Error("Stderr not set correctly")
	}

	if runner.Client == nil {
		t.Error("Client should not be nil")
	}
}

func TestRunFile_EmptyPath(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	runner := NewRunner(stdout, stderr)

	_, err := runner.RunFile("", "https://example.com")
	if err == nil {
		t.Error("Expected error for empty path")
	}

	if !strings.Contains(err.Error(), "input path is required") {
		t.Errorf("Expected 'input path is required' error, got: %v", err)
	}
}

func TestRunFile_EmptyGradioURL(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	runner := NewRunner(stdout, stderr)

	tmpFile := filepath.Join(t.TempDir(), "test.json")
	os.WriteFile(tmpFile, []byte("{}"), 0644)

	_, err := runner.RunFile(tmpFile, "")
	if err == nil {
		t.Error("Expected error for empty gradio URL")
	}

	if !strings.Contains(err.Error(), "gradio URL is required") {
		t.Errorf("Expected 'gradio URL is required' error, got: %v", err)
	}
}

func TestRunFile_InvalidJSON(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	runner := NewRunner(stdout, stderr)

	tmpFile := filepath.Join(t.TempDir(), "invalid.json")
	os.WriteFile(tmpFile, []byte("invalid json"), 0644)

	_, err := runner.RunFile(tmpFile, "https://example.com")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestRunFile_ValidJSON(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	runner := NewRunner(stdout, stderr)

	tmpFile := filepath.Join(t.TempDir(), "test.json")
	tmpDir := t.TempDir()

	input := InputData{
		FolderName: tmpDir,
		Items:      []ImageRequest{},
	}

	data, _ := json.Marshal(input)
	os.WriteFile(tmpFile, data, 0644)

	// Mock Gradio server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"event_id": "test-event-id",
		})
	}))
	defer server.Close()

	results, err := runner.RunFile(tmpFile, server.URL)
	if err != nil {
		t.Logf("RunFile with empty items returned: %v (expected)", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty items, got %d", len(results))
	}
}

func TestGenerate_EmptyItems(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	runner := NewRunner(stdout, stderr)

	input := InputData{
		FolderName: t.TempDir(),
		Items:      []ImageRequest{},
	}

	results := runner.Generate(input, "https://example.com")

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty items, got %d", len(results))
	}
}

func TestImageRequestStructure(t *testing.T) {
	req := ImageRequest{
		ID:                "IMAGE_1",
		Alt:               "Test",
		Prompt:            "Test prompt",
		Style:             "photorealistic",
		AspectRatio:       "16:9",
		PositionInArticle: 100,
		Platform:          "web",
	}

	if req.ID != "IMAGE_1" {
		t.Errorf("ID = %q, want %q", req.ID, "IMAGE_1")
	}

	if req.Platform != "web" {
		t.Errorf("Platform = %q, want %q", req.Platform, "web")
	}
}

func TestGradioRequestStructure(t *testing.T) {
	req := GradioRequest{
		Data: []interface{}{"test", 1920, 1080},
	}

	if len(req.Data) != 3 {
		t.Errorf("Expected 3 data items, got %d", len(req.Data))
	}
}

func TestImageOutputMetadataStructure(t *testing.T) {
	seed := int64(12345)
	metadata := ImageOutputMetadata{
		Request: ImageRequest{
			ID:     "IMAGE_1",
			Prompt: "Test",
		},
		GradioPayload: GradioPayload{
			Width:  1920,
			Height: 1080,
		},
		Seed:        &seed,
		DurationMs:  1000,
		EventID:     "test-event",
		GeneratedAt: "2024-01-01T00:00:00Z",
	}

	if metadata.EventID != "test-event" {
		t.Errorf("EventID = %q, want %q", metadata.EventID, "test-event")
	}

	if *metadata.Seed != 12345 {
		t.Errorf("Seed = %d, want %d", *metadata.Seed, 12345)
	}
}

func TestResultStructure(t *testing.T) {
	req := ImageRequest{ID: "IMAGE_1"}

	// Test with no error
	result := Result{
		Request: req,
		Err:     nil,
	}

	if result.Err != nil {
		t.Error("Expected nil error")
	}

	// Test with error
	result2 := Result{
		Request: req,
		Err:     os.ErrNotExist,
	}

	if result2.Err == nil {
		t.Error("Expected non-nil error")
	}
}
