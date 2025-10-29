package serp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FetchResults pobiera wyniki SERP z instancji SearXNG.
func FetchResults(searxngURL, query, lang, country string, limit int) ([]Result, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("language", lang)
	params.Add("format", "json")
	params.Add("safesearch", "0")

	fullURL := fmt.Sprintf("%s/search?%s", strings.TrimRight(searxngURL, "/"), params.Encode())

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BriefBot/1.0)")
	req.Header.Set("Accept", "application/json")

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

	var searxResponse struct {
		Results []struct {
			URL     string `json:"url"`
			Title   string `json:"title"`
			Content string `json:"content"`
			Engine  string `json:"engine"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &searxResponse); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	results := make([]Result, 0, limit)
	position := 1

	for _, item := range searxResponse.Results {
		if len(results) >= limit {
			break
		}

		if item.URL == "" || item.Title == "" {
			continue
		}

		parsedURL, err := url.Parse(item.URL)
		host := ""
		if err == nil {
			host = parsedURL.Host
		}

		results = append(results, Result{
			Position: position,
			URL:      item.URL,
			Title:    item.Title,
			Snippet:  item.Content,
			Host:     host,
		})
		position++
	}

	return results, nil
}

// NewResponse buduje strukturę odpowiedzi wraz ze znacznikiem czasu.
func NewResponse(query, lang, country string, results []Result) Response {
	return Response{
		Query:     query,
		Language:  lang,
		Country:   country,
		Results:   results,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}
