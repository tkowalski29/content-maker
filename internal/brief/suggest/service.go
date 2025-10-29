package suggest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FetchSuggestions pobiera podpowiedzi z wybranego dostawcy.
func FetchSuggestions(provider, query, lang, country string) ([]string, error) {
	switch provider {
	case "google":
		return fetchGoogle(query, lang, country)
	case "bing":
		return fetchBing(query, lang, country)
	case "dgd":
		return fetchDuckDuckGo(query, lang)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func fetchGoogle(query, lang, country string) ([]string, error) {
	baseURL := "http://suggestqueries.google.com/complete/search"
	params := url.Values{}
	params.Add("client", "firefox")
	params.Add("q", query)
	params.Add("hl", lang)
	params.Add("gl", country)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fullURL)
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

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if len(result) < 2 {
		return []string{}, nil
	}

	raw, ok := result[1].([]interface{})
	if !ok {
		return []string{}, nil
	}

	suggestions := make([]string, 0, len(raw))
	for _, item := range raw {
		if str, ok := item.(string); ok {
			suggestions = append(suggestions, str)
		}
	}

	return suggestions, nil
}

func fetchBing(query, lang, country string) ([]string, error) {
	baseURL := "https://www.bing.com/AS/Suggestions"
	params := url.Values{}
	params.Add("q", query)
	params.Add("mkt", fmt.Sprintf("%s-%s", lang, country))
	params.Add("qry", query)
	params.Add("cvid", "0")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fullURL, nil)
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

	var result struct {
		AS struct {
			Results []struct {
				Suggests []struct {
					Txt string `json:"Txt"`
				} `json:"Suggests"`
			} `json:"Results"`
		} `json:"AS"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	suggestions := make([]string, 0)
	for _, r := range result.AS.Results {
		for _, s := range r.Suggests {
			if s.Txt != "" {
				suggestions = append(suggestions, s.Txt)
			}
		}
	}

	return suggestions, nil
}

func fetchDuckDuckGo(query, lang string) ([]string, error) {
	baseURL := "https://duckduckgo.com/ac/"
	params := url.Values{}
	params.Add("q", query)
	params.Add("kl", fmt.Sprintf("%s-xx", lang))

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fullURL, nil)
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

	var result []struct {
		Phrase string `json:"phrase"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	suggestions := make([]string, 0, len(result))
	for _, item := range result {
		if item.Phrase != "" && !strings.EqualFold(item.Phrase, query) {
			suggestions = append(suggestions, item.Phrase)
		}
	}

	return suggestions, nil
}

// Response opisuje dane zwracane do użytkownika CLI.
type Response struct {
	Provider    string   `json:"provider"`
	Query       string   `json:"query"`
	Suggestions []string `json:"suggestions"`
	Timestamp   string   `json:"timestamp"`
}

// NewResponse buduje strukturę odpowiedzi.
func NewResponse(provider, query string, suggestions []string) Response {
	return Response{
		Provider:    provider,
		Query:       query,
		Suggestions: suggestions,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
}
