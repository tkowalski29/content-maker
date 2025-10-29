package fetch

// HTMLResponse reprezentuje wynik pobierania i parsowania HTML.
type HTMLResponse struct {
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Text        string   `json:"text"`
	H1          []string `json:"h1"`
	H2          []string `json:"h2"`
	H3          []string `json:"h3"`
	WordCount   int      `json:"word_count"`
	FetchedAt   string   `json:"fetched_at"`
	FetchMethod string   `json:"fetch_method"`
}
