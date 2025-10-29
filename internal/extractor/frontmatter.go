package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontmatterOptions configures frontmatter extraction.
type FrontmatterOptions struct {
	InputPath  string
	OutputPath string
}

// ExtractFrontmatter parses frontmatter and writes JSON to the given path.
func ExtractFrontmatter(opts FrontmatterOptions) error {
	if opts.InputPath == "" || opts.OutputPath == "" {
		return fmt.Errorf("input and output paths are required")
	}

	content, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	cmsData, err := ExtractCMSData(string(content))
	if err != nil {
		return fmt.Errorf("extracting CMS data: %w", err)
	}

	jsonData, err := json.MarshalIndent(cmsData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(opts.OutputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// ExtractCMSData wyciąga strukturę CMS z treści markdown.
func ExtractCMSData(content string) (*CMSData, error) {
	frontmatterRe := regexp.MustCompile(`(?s)^---\n(.*?)\n---`)
	matches := frontmatterRe.FindStringSubmatch(content)

	if len(matches) < 2 {
		return nil, fmt.Errorf("no front-matter found")
	}

	yamlContent := matches[1]

	var cmsData CMSData
	if err := yaml.Unmarshal([]byte(yamlContent), &cmsData); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	if cmsData.Slug == "" {
		cmsData.Slug = GenerateSlug(cmsData.Title)
	}

	faq := extractFAQ(content)
	if len(faq) > 0 {
		cmsData.SchemaOrg = SchemaOrg{
			Type:        "Article",
			Headline:    cmsData.Title,
			Description: cmsData.Description,
			FAQ:         faq,
		}
	}

	return &cmsData, nil
}

func extractFAQ(content string) []FAQItem {
	var faq []FAQItem

	faqRe := regexp.MustCompile(`(?i)##\s*(FAQ|Najczęściej\s+zadawane\s+pytania).*?\n([\s\S]*?)(?:\n##\s|\z)`)
	matches := faqRe.FindStringSubmatch(content)

	if len(matches) < 3 {
		return faq
	}

	faqSection := matches[2]
	lines := strings.Split(faqSection, "\n")

	var (
		questions       [][]string
		currentQuestion string
		currentAnswer   strings.Builder
	)

	for _, line := range lines {
		if strings.HasPrefix(line, "###") {
			if currentQuestion != "" {
				questions = append(questions, []string{"", currentQuestion, currentAnswer.String()})
			}
			currentQuestion = strings.TrimSpace(strings.TrimPrefix(line, "###"))
			currentAnswer.Reset()
		} else if currentQuestion != "" && !strings.HasPrefix(line, "##") {
			if currentAnswer.Len() > 0 {
				currentAnswer.WriteString("\n")
			}
			currentAnswer.WriteString(line)
		}
	}

	if currentQuestion != "" {
		questions = append(questions, []string{"", currentQuestion, currentAnswer.String()})
	}

	for _, q := range questions {
		if len(q) < 3 {
			continue
		}

		question := strings.TrimSpace(q[1])
		answer := strings.TrimSpace(q[2])

		answer = cleanMarkdown(answer)

		faq = append(faq, FAQItem{
			Question: question,
			Answer:   answer,
		})
	}

	return faq
}

func cleanMarkdown(text string) string {
	text = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`[*_]{1,2}([^*_]+)[*_]{1,2}`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "$1")
	text = strings.ReplaceAll(text, "\n\n", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	return strings.TrimSpace(text)
}

// GenerateSlug tworzy przyjazny adres URL.
func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if len(slug) > 100 {
		slug = slug[:100]
		if lastDash := strings.LastIndex(slug, "-"); lastDash > 0 {
			slug = slug[:lastDash]
		}
	}

	return slug
}
