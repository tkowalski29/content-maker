package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FetchHTML pobiera i przetwarza stronę HTML zgodnie z wybraną metodą.
func FetchHTML(targetURL, method string) (*HTMLResponse, error) {
	switch method {
	case "netlify":
		return fetchWithNetlify(targetURL)
	case "jina":
		return fetchWithJina(targetURL)
	case "direct":
		return fetchDirect(targetURL)
	default:
		return nil, fmt.Errorf("unknown method: %s (available: netlify, jina, direct)", method)
	}
}

func fetchWithJina(targetURL string) (*HTMLResponse, error) {
	jinaURL := fmt.Sprintf("https://r.jina.ai/%s", targetURL)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", jinaURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Return-Format", "json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var jinaResponse struct {
		Data struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			URL     string `json:"url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &jinaResponse); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	h1, h2, h3 := extractHeadings(jinaResponse.Data.Content)

	result := &HTMLResponse{
		URL:         targetURL,
		Title:       jinaResponse.Data.Title,
		Text:        jinaResponse.Data.Content,
		H1:          h1,
		H2:          h2,
		H3:          h3,
		WordCount:   countWords(jinaResponse.Data.Content),
		FetchedAt:   time.Now().UTC().Format(time.RFC3339),
		FetchMethod: "jina",
	}

	return result, nil
}

func fetchWithNetlify(targetURL string) (*HTMLResponse, error) {
	apiURL := "https://"

	requestBody := map[string]string{"url": targetURL}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var netlifyResponse struct {
		OK       bool   `json:"ok"`
		URL      string `json:"url"`
		Title    string `json:"title"`
		Headings struct {
			H2 []string `json:"h2"`
			H3 []string `json:"h3"`
		} `json:"headings"`
		Text               string `json:"text"`
		ContentLengthWords int    `json:"content_length_words"`
		MetaDescription    string `json:"metaDescription"`
		SourceMode         string `json:"source_mode"`
		NeedFallback       bool   `json:"needFallback"`
	}

	if err := json.Unmarshal(body, &netlifyResponse); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if !netlifyResponse.OK {
		return nil, fmt.Errorf("extractor returned ok=false")
	}

	return &HTMLResponse{
		URL:         targetURL,
		Title:       netlifyResponse.Title,
		Text:        netlifyResponse.Text,
		H1:          []string{},
		H2:          netlifyResponse.Headings.H2,
		H3:          netlifyResponse.Headings.H3,
		WordCount:   netlifyResponse.ContentLengthWords,
		FetchedAt:   time.Now().UTC().Format(time.RFC3339),
		FetchMethod: "netlify",
	}, nil
}

func fetchDirect(targetURL string) (*HTMLResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BriefBot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	html := string(body)

	title := extractTitle(html)
	text := stripTags(html)
	h1, h2, h3 := extractHeadingsFromHTML(html)

	return &HTMLResponse{
		URL:         targetURL,
		Title:       title,
		Text:        text,
		H1:          h1,
		H2:          h2,
		H3:          h3,
		WordCount:   countWords(text),
		FetchedAt:   time.Now().UTC().Format(time.RFC3339),
		FetchMethod: "direct",
	}, nil
}

func extractHeadings(content string) ([]string, []string, []string) {
	lines := strings.Split(content, "\n")
	var h1, h2, h3 []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "# "):
			h1 = append(h1, strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "## "):
			h2 = append(h2, strings.TrimPrefix(line, "## "))
		case strings.HasPrefix(line, "### "):
			h3 = append(h3, strings.TrimPrefix(line, "### "))
		}
	}

	return h1, h2, h3
}

func extractHeadingsFromHTML(html string) ([]string, []string, []string) {
	// Wersja uproszczona – w produkcji użyj parsera HTML.
	return []string{}, []string{}, []string{}
}

func extractTitle(html string) string {
	start := strings.Index(html, "<title>")
	end := strings.Index(html, "</title>")
	if start != -1 && end != -1 && end > start {
		return html[start+7 : end]
	}
	return ""
}

func stripTags(html string) string {
	text := html

	for {
		start := strings.Index(text, "<script")
		end := strings.Index(text, "</script>")
		if start == -1 || end == -1 {
			break
		}
		text = text[:start] + text[end+9:]
	}

	for {
		start := strings.Index(text, "<style")
		end := strings.Index(text, "</style>")
		if start == -1 || end == -1 {
			break
		}
		text = text[:start] + text[end+8:]
	}

	text = strings.ReplaceAll(text, "<", "\n<")
	text = strings.ReplaceAll(text, ">", ">\n")

	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		if !strings.Contains(line, "<") && !strings.Contains(line, ">") {
			line = strings.TrimSpace(line)
			if line != "" {
				cleaned = append(cleaned, line)
			}
		}
	}

	return strings.Join(cleaned, " ")
}

func countWords(text string) int {
	return len(strings.Fields(text))
}
