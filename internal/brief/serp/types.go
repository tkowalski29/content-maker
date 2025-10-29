package serp

// Result opisuje pojedynczy wynik SERP.
type Result struct {
	Position int    `json:"position"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Snippet  string `json:"snippet"`
	Host     string `json:"host"`
}

// Response zawiera kompletną odpowiedź SERP.
type Response struct {
	Query     string   `json:"query"`
	Language  string   `json:"language"`
	Country   string   `json:"country"`
	Results   []Result `json:"results"`
	Timestamp string   `json:"timestamp"`
}
