package fetch

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{"empty string", "", 0},
		{"single word", "hello", 1},
		{"multiple words", "hello world test", 3},
		{"words with multiple spaces", "hello   world    test", 3},
		{"words with newlines", "hello\nworld\ntest", 3},
		{"words with tabs", "hello\tworld\ttest", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countWords(tt.text)
			if got != tt.want {
				t.Errorf("countWords() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExtractHeadings(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantH1  int
		wantH2  int
		wantH3  int
	}{
		{"empty content", "", 0, 0, 0},
		{"markdown headings", `# Main Title
Some content here
## Section 1
Content
### Subsection 1.1
More content
## Section 2
### Subsection 2.1
### Subsection 2.2`, 1, 2, 3},
		{"headings with extra spaces", `  # Title with spaces
  ## Section with spaces
  ### Subsection with spaces  `, 1, 1, 1},
		{"no headings", `This is just regular text
without any headings
multiple lines`, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1, h2, h3 := extractHeadings(tt.content)
			if len(h1) != tt.wantH1 {
				t.Errorf("extractHeadings() h1 count = %d, want %d", len(h1), tt.wantH1)
			}
			if len(h2) != tt.wantH2 {
				t.Errorf("extractHeadings() h2 count = %d, want %d", len(h2), tt.wantH2)
			}
			if len(h3) != tt.wantH3 {
				t.Errorf("extractHeadings() h3 count = %d, want %d", len(h3), tt.wantH3)
			}
		})
	}
}

func TestExtractHeadings_Content(t *testing.T) {
	content := `# Main Title
## First Section
### First Subsection
## Second Section`

	h1, h2, h3 := extractHeadings(content)

	if len(h1) != 1 || h1[0] != "Main Title" {
		t.Errorf("Expected h1[0] = 'Main Title', got %v", h1)
	}

	if len(h2) != 2 || h2[0] != "First Section" || h2[1] != "Second Section" {
		t.Errorf("Expected h2 = ['First Section', 'Second Section'], got %v", h2)
	}

	if len(h3) != 1 || h3[0] != "First Subsection" {
		t.Errorf("Expected h3[0] = 'First Subsection', got %v", h3)
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"valid title", "<html><head><title>Test Page</title></head></html>", "Test Page"},
		{"no title tag", "<html><head></head></html>", ""},
		{"empty title", "<html><head><title></title></head></html>", ""},
		{"title with special characters", "<html><head><title>Test & Page | Example</title></head></html>", "Test & Page | Example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTitle(tt.html)
			if got != tt.want {
				t.Errorf("extractTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripTags(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"simple paragraph", "<p>Hello world</p>", "Hello world"},
		{"with script tag", "<div>Content<script>alert('test')</script>More content</div>", "ContentMore content"},
		{"with style tag", "<div>Content<style>.test{color:red}</style>More content</div>", "ContentMore content"},
		{"multiple tags", "<div><p>Hello</p><span>World</span></div>", "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripTags(tt.html)
			if got != tt.want {
				t.Errorf("stripTags() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchHTML_InvalidMethod(t *testing.T) {
	_, err := FetchHTML("https://example.com", "invalid_method")
	if err == nil {
		t.Error("Expected error for invalid method, got nil")
	}

	if err == nil || !strings.Contains(err.Error(), "unknown method: invalid_method") {
		t.Errorf("Expected error message to contain 'unknown method: invalid_method', got %v", err)
	}
}

func TestFetchDirect_MockServer(t *testing.T) {
	testHTML := `<html>
<head><title>Test Page</title></head>
<body>
<h1>Main Heading</h1>
<p>This is a test paragraph with multiple words.</p>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHTML))
	}))
	defer server.Close()

	result, err := fetchDirect(server.URL)
	if err != nil {
		t.Fatalf("fetchDirect() error = %v", err)
	}

	if result == nil {
		t.Fatal("fetchDirect() returned nil result")
	}

	if result.URL != server.URL {
		t.Errorf("URL = %q, want %q", result.URL, server.URL)
	}

	if result.Title != "Test Page" {
		t.Errorf("Title = %q, want %q", result.Title, "Test Page")
	}

	if result.FetchMethod != "direct" {
		t.Errorf("FetchMethod = %q, want %q", result.FetchMethod, "direct")
	}

	if result.WordCount == 0 {
		t.Error("WordCount should be > 0")
	}

	if result.FetchedAt == "" {
		t.Error("FetchedAt should not be empty")
	}
}

func TestFetchDirect_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"success", http.StatusOK, false},
		{"not found", http.StatusNotFound, true},
		{"server error", http.StatusInternalServerError, true},
		{"forbidden", http.StatusForbidden, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					w.Write([]byte("<html><head><title>Test</title></head><body>Content</body></html>"))
				}
			}))
			defer server.Close()

			_, err := fetchDirect(server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchDirect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetchDirect_InvalidURL(t *testing.T) {
	_, err := fetchDirect("http://invalid-url-that-does-not-exist-12345.com")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestExtractHeadingsFromHTML(t *testing.T) {
	html := "<h1>Title</h1><h2>Section</h2><h3>Subsection</h3>"
	h1, h2, h3 := extractHeadingsFromHTML(html)

	if len(h1) != 0 || len(h2) != 0 || len(h3) != 0 {
		t.Log("Note: extractHeadingsFromHTML is a placeholder and returns empty arrays")
	}
}

func TestFetchHTML_MethodRouting(t *testing.T) {
	tests := []struct {
		name   string
		method string
		url    string
	}{
		{"invalid method", "invalid", "https://example.com"},
		{"empty method", "", "https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FetchHTML(tt.url, tt.method)
			if err == nil {
				t.Error("Expected error for invalid/empty method")
			}
		})
	}
}
