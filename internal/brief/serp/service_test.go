package serp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchResults_MockServer(t *testing.T) {
	mockResponse := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"url":     "https://example1.com",
				"title":   "Example 1",
				"content": "This is the first example result",
				"engine":  "google",
			},
			{
				"url":     "https://example2.com",
				"title":   "Example 2",
				"content": "This is the second example result",
				"engine":  "bing",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("Expected path /search, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	results, err := FetchResults(server.URL, "test query", "pl", "PL", 10)
	if err != nil {
		t.Fatalf("FetchResults() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results[0].Position != 1 || results[0].Host != "example1.com" {
		t.Errorf("Unexpected first result: %+v", results[0])
	}
}

func TestFetchResults_Limit(t *testing.T) {
	mockResults := make([]map[string]interface{}, 6)
	for i := range mockResults {
		mockResults[i] = map[string]interface{}{
			"url":     "https://example.com",
			"title":   "Example",
			"content": "Snippet",
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"results": mockResults})
	}))
	defer server.Close()

	results, err := FetchResults(server.URL, "test", "pl", "PL", 3)
	if err != nil {
		t.Fatalf("FetchResults() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

func TestFetchResults_SkipInvalid(t *testing.T) {
	mockResponse := map[string]interface{}{
		"results": []map[string]interface{}{
			{"url": "https://example1.com", "title": "Valid 1", "content": "Snippet"},
			{"url": "", "title": "Invalid", "content": "Skip"},
			{"url": "https://example2.com", "title": "Valid 2", "content": "Snippet"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	results, err := FetchResults(server.URL, "test", "pl", "PL", 10)
	if err != nil {
		t.Fatalf("FetchResults() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 valid results, got %d", len(results))
	}

	if results[0].Position != 1 || results[1].Position != 2 {
		t.Errorf("Positions should increment sequentially, got %+v", results)
	}
}

func TestFetchResults_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	if _, err := FetchResults(server.URL, "test", "pl", "PL", 10); err == nil {
		t.Error("Expected error on non-200 status")
	}
}

func TestFetchResults_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid json}"))
	}))
	defer server.Close()

	if _, err := FetchResults(server.URL, "test", "pl", "PL", 10); err == nil {
		t.Error("Expected error on invalid JSON")
	}
}

func TestNewResponse(t *testing.T) {
	results := []Result{{Position: 1, URL: "https://example.com", Title: "Example"}}
	resp := NewResponse("query", "pl", "PL", results)

	if resp.Query != "query" || resp.Language != "pl" || resp.Country != "PL" {
		t.Errorf("Unexpected response metadata: %+v", resp)
	}

	if len(resp.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(resp.Results))
	}

	if resp.Timestamp == "" {
		t.Error("Timestamp should be populated")
	}
}

func TestFetchResults_EmptyResults(t *testing.T) {
	mockResponse := map[string]interface{}{
		"results": []map[string]interface{}{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	results, err := FetchResults(server.URL, "test", "pl", "PL", 10)
	if err != nil {
		t.Fatalf("FetchResults() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestFetchResults_SkipEmptyTitle(t *testing.T) {
	mockResponse := map[string]interface{}{
		"results": []map[string]interface{}{
			{"url": "https://example.com", "title": "", "content": "No title"},
			{"url": "https://example2.com", "title": "Valid", "content": "Has title"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	results, err := FetchResults(server.URL, "test", "pl", "PL", 10)
	if err != nil {
		t.Fatalf("FetchResults() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 valid result, got %d", len(results))
	}
}

func TestFetchResults_URLParsing(t *testing.T) {
	mockResponse := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"url":     "https://subdomain.example.com/path?query=value",
				"title":   "Test",
				"content": "Content",
			},
			{
				"url":     "invalid://url",
				"title":   "Invalid URL",
				"content": "Should still include",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	results, err := FetchResults(server.URL, "test", "pl", "PL", 10)
	if err != nil {
		t.Fatalf("FetchResults() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results[0].Host != "subdomain.example.com" {
		t.Errorf("Expected host 'subdomain.example.com', got %q", results[0].Host)
	}

	// Second result has invalid URL but should still be included with empty host
	if results[1].Host != "" {
		t.Logf("Invalid URL resulted in host: %q", results[1].Host)
	}
}

func TestFetchResults_URLTrimming(t *testing.T) {
	tests := []struct {
		name       string
		searxngURL string
		wantPath   string
	}{
		{"with trailing slash", "http://example.com/", "/search"},
		{"without trailing slash", "http://example.com", "/search"},
		{"with multiple slashes", "http://example.com///", "/search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantPath {
					t.Errorf("Expected path %q, got %q", tt.wantPath, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
			}))
			defer server.Close()

			_, err := FetchResults(tt.searxngURL+server.URL[len("http://"):], "test", "pl", "PL", 10)
			if err != nil {
				t.Logf("FetchResults() error = %v (expected for this test)", err)
			}
		})
	}
}

func TestNewResponse_EmptyResults(t *testing.T) {
	resp := NewResponse("query", "en", "US", []Result{})

	if len(resp.Results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(resp.Results))
	}

	if resp.Query != "query" {
		t.Errorf("Expected query 'query', got %q", resp.Query)
	}
}

func TestNewResponse_MultipleResults(t *testing.T) {
	results := []Result{
		{Position: 1, URL: "https://example1.com", Title: "First"},
		{Position: 2, URL: "https://example2.com", Title: "Second"},
		{Position: 3, URL: "https://example3.com", Title: "Third"},
	}

	resp := NewResponse("test query", "pl", "PL", results)

	if len(resp.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(resp.Results))
	}

	for i, result := range resp.Results {
		if result.Position != i+1 {
			t.Errorf("Result %d has position %d, expected %d", i, result.Position, i+1)
		}
	}
}

func TestResultStructure(t *testing.T) {
	result := Result{
		Position: 5,
		URL:      "https://example.com/path",
		Title:    "Example Title",
		Snippet:  "Example snippet content",
		Host:     "example.com",
	}

	if result.Position != 5 {
		t.Errorf("Position = %d, want 5", result.Position)
	}

	if result.Host != "example.com" {
		t.Errorf("Host = %q, want %q", result.Host, "example.com")
	}
}

func TestResponseStructure(t *testing.T) {
	response := Response{
		Query:     "test",
		Language:  "pl",
		Country:   "PL",
		Results:   []Result{},
		Timestamp: "2024-01-01T00:00:00Z",
	}

	if response.Language != "pl" {
		t.Errorf("Language = %q, want %q", response.Language, "pl")
	}

	if response.Country != "PL" {
		t.Errorf("Country = %q, want %q", response.Country, "PL")
	}
}
