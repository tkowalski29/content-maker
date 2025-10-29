package extractor

// Image reprezentuje dane obrazka wyciągniętego z markdown.
type Image struct {
	ID                string `json:"id"`
	Alt               string `json:"alt"`
	Prompt            string `json:"prompt"`
	Style             string `json:"style"`
	AspectRatio       string `json:"aspect_ratio"`
	PositionInArticle int    `json:"position_in_article"`
}

// CMSData reprezentuje metadane CMS wczytane z front-matter.
type CMSData struct {
	Title         string    `json:"title" yaml:"title"`
	Slug          string    `json:"slug" yaml:"slug"`
	Description   string    `json:"description" yaml:"description"`
	Keywords      []string  `json:"keywords" yaml:"keywords"`
	OGTitle       string    `json:"og_title" yaml:"og_title"`
	OGDescription string    `json:"og_description" yaml:"og_description"`
	OGImage       string    `json:"og_image" yaml:"og_image"`
	OGType        string    `json:"og_type" yaml:"og_type"`
	TwitterCard   string    `json:"twitter_card" yaml:"twitter_card"`
	Lang          string    `json:"lang" yaml:"lang"`
	Author        string    `json:"author" yaml:"author"`
	PublishedAt   string    `json:"published_at" yaml:"published_at"`
	ModifiedAt    string    `json:"modified_at" yaml:"modified_at"`
	Canonical     string    `json:"canonical" yaml:"canonical"`
	SchemaOrg     SchemaOrg `json:"schema_org,omitempty"`
}

// SchemaOrg opisuje dane schema.org.
type SchemaOrg struct {
	Type        string    `json:"type"`
	Headline    string    `json:"headline"`
	Description string    `json:"description"`
	FAQ         []FAQItem `json:"faq,omitempty"`
}

// FAQItem opisuje pojedyncze pytanie i odpowiedź.
type FAQItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}
