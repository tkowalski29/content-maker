package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ImageOptions configures image extraction.
type ImageOptions struct {
	InputPath  string
	OutputPath string
}

// ExtractImages reads markdown, extracts image placeholders, and writes JSON.
func ExtractImages(opts ImageOptions) (int, error) {
	if opts.InputPath == "" || opts.OutputPath == "" {
		return 0, fmt.Errorf("input and output paths are required")
	}

	content, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return 0, fmt.Errorf("reading file: %w", err)
	}

	images := ParseImages(string(content))

	jsonData, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(opts.OutputPath, jsonData, 0644); err != nil {
		return 0, fmt.Errorf("writing file: %w", err)
	}

	return len(images), nil
}

// ParseImages zwraca listę obrazów znalezionych w treści markdown.
func ParseImages(content string) []Image {
	var images []Image

	placeholderRe := regexp.MustCompile(`\{\{IMAGE_(\d+)\}\}`)
	commentRe := regexp.MustCompile(`<!--\s*IMAGE_(\d+):\s*alt="([^"]+)",\s*prompt="([^"]+)"(?:,\s*style="([^"]+)")?(?:,\s*aspect="([^"]+)")?\s*-->`)

	placeholders := placeholderRe.FindAllStringSubmatch(content, -1)
	comments := commentRe.FindAllStringSubmatch(content, -1)

	commentMap := make(map[string][]string)
	for _, match := range comments {
		if len(match) >= 4 {
			id := match[1]
			alt := match[2]
			prompt := match[3]
			style := "photorealistic"
			aspect := "16:9"

			if len(match) > 4 && match[4] != "" {
				style = match[4]
			}
			if len(match) > 5 && match[5] != "" {
				aspect = match[5]
			}

			commentMap[id] = []string{alt, prompt, style, aspect}
		}
	}

	for _, match := range placeholders {
		if len(match) < 2 {
			continue
		}

		id := match[1]
		imageID := "IMAGE_" + id

		alt := ""
		prompt := ""
		style := "photorealistic"
		aspect := "16:9"

		if metadata, ok := commentMap[id]; ok && len(metadata) >= 2 {
			alt = metadata[0]
			prompt = metadata[1]
			if len(metadata) > 2 {
				style = metadata[2]
			}
			if len(metadata) > 3 {
				aspect = metadata[3]
			}
		}

		position := strings.Index(content, "{{IMAGE_"+id+"}}")

		images = append(images, Image{
			ID:                imageID,
			Alt:               alt,
			Prompt:            prompt,
			Style:             style,
			AspectRatio:       aspect,
			PositionInArticle: position,
		})
	}

	return images
}
