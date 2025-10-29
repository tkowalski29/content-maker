package analyze

// Audit represents pojedynczy audyt adresu URL.
type Audit struct {
	URL            string         `json:"url"`
	MissingTopics  []string       `json:"missing_topics"`
	EntityCoverage []EntityStatus `json:"entity_coverage"`
	UniqueAngles   []string       `json:"unique_angles"`
	QualityFlags   []string       `json:"quality_flags"`
	ContentLength  int            `json:"content_length_words"`
}

// EntityStatus opisuje pokrycie encji.
type EntityStatus struct {
	Entity string `json:"entity"`
	Status string `json:"status"`
}

// Analysis zawiera zagregowane wyniki analizy.
type Analysis struct {
	ContentGaps       []Gap            `json:"content_gaps"`
	EntitiesToCover   []EntityAnalysis `json:"entities_to_cover"`
	EntitiesToImprove []EntityAnalysis `json:"entities_to_improve"`
	CompetitorAngles  []string         `json:"competitor_unique_angles"`
	AverageWordCount  float64          `json:"avg_word_count"`
	TopQualityIssues  []QualityIssue   `json:"top_quality_issues"`
}

// Gap reprezentuje lukę tematyczną.
type Gap struct {
	Topic        string `json:"topic"`
	MissingCount int    `json:"missing_count"`
}

// EntityAnalysis reprezentuje zagregowane dane o encji.
type EntityAnalysis struct {
	Entity      string `json:"entity"`
	Status      string `json:"status"`
	Occurrences int    `json:"occurrences"`
}

// QualityIssue opisuje najczęstsze problemy jakościowe.
type QualityIssue struct {
	Issue string `json:"issue"`
	Count int    `json:"count"`
}
