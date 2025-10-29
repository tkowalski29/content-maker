package extractor

import "testing"

func TestParseImages(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
		checkID  string
		checkAlt string
	}{
		{
			name: "single image with metadata",
			content: `# Article

{{IMAGE_1}}
<!-- IMAGE_1: alt="Test image", prompt="A test image prompt", style="photorealistic", aspect="16:9" -->

Some content.`,
			expected: 1,
			checkID:  "IMAGE_1",
			checkAlt: "Test image",
		},
		{
			name: "image without metadata",
			content: `# Article

{{IMAGE_1}}

No metadata comment.`,
			expected: 1,
			checkID:  "IMAGE_1",
			checkAlt: "",
		},
		{
			name:     "no images",
			content:  "# Article\n\nJust text content.",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			images := ParseImages(tt.content)
			if len(images) != tt.expected {
				t.Fatalf("expected %d images, got %d", tt.expected, len(images))
			}
			if tt.expected > 0 && tt.checkID != "" {
				found := false
				for _, img := range images {
					if img.ID == tt.checkID {
						found = true
						if img.Alt != tt.checkAlt {
							t.Errorf("expected alt %q, got %q", tt.checkAlt, img.Alt)
						}
					}
				}
				if !found {
					t.Errorf("image with ID %s not found", tt.checkID)
				}
			}
		})
	}
}

func TestParseImagesDefaults(t *testing.T) {
	content := `{{IMAGE_1}}
<!-- IMAGE_1: alt="Test", prompt="Prompt" -->`
	images := ParseImages(content)
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	if images[0].Style != "photorealistic" {
		t.Errorf("expected default style, got %s", images[0].Style)
	}
	if images[0].AspectRatio != "16:9" {
		t.Errorf("expected default aspect ratio, got %s", images[0].AspectRatio)
	}
}

func TestParseImagesPosition(t *testing.T) {
	content := `Intro text.

{{IMAGE_1}}
<!-- IMAGE_1: alt="Test", prompt="Prompt" -->`
	images := ParseImages(content)
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	if images[0].PositionInArticle <= 0 {
		t.Error("expected positive position in article")
	}
}

func TestExtractCMSData(t *testing.T) {
	content := `---
title: "Test Article"
description: "Test description"
keywords:
  - test
  - article
lang: "pl"
author: "Test Author"
---

# Test Article

## FAQ

### What is this?

This is a *test* article.

### How does it work?

It works **great**.`

	cms, err := ExtractCMSData(content)
	if err != nil {
		t.Fatalf("ExtractCMSData error: %v", err)
	}

	if cms.Title != "Test Article" {
		t.Errorf("expected title, got %s", cms.Title)
	}
	if cms.Slug != "test-article" {
		t.Errorf("expected slug 'test-article', got %s", cms.Slug)
	}
	if len(cms.SchemaOrg.FAQ) != 2 {
		t.Fatalf("expected 2 FAQ items, got %d", len(cms.SchemaOrg.FAQ))
	}
	if cms.SchemaOrg.FAQ[0].Answer == "" {
		t.Error("expected FAQ answer to be cleaned")
	}
}

func TestExtractCMSData_NoFrontmatter(t *testing.T) {
	if _, err := ExtractCMSData("# Article without frontmatter"); err == nil {
		t.Fatal("expected error when frontmatter missing")
	}
}

func TestGenerateSlug(t *testing.T) {
	if slug := GenerateSlug("This is a Test Title!"); slug != "this-is-a-test-title" {
		t.Errorf("unexpected slug: %s", slug)
	}

	longTitle := "This is a very long title that should be truncated at exactly one hundred characters to ensure URLs stay manageable"
	slug := GenerateSlug(longTitle)
	if len(slug) > 100 {
		t.Errorf("slug should be trimmed to max 100 characters, got %d", len(slug))
	}
}
