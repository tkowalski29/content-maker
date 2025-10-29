package analyze

import "testing"

func TestAnalyzeAudits(t *testing.T) {
	tests := []struct {
		name    string
		audits  []Audit
		wantLen struct {
			gaps            int
			entitiesToCover int
			entitiesToImpr  int
			angles          int
			qualityIssues   int
		}
		wantAvgWords float64
	}{
		{
			name:   "empty audits",
			audits: []Audit{},
			wantLen: struct {
				gaps            int
				entitiesToCover int
				entitiesToImpr  int
				angles          int
				qualityIssues   int
			}{0, 0, 0, 0, 0},
			wantAvgWords: 0.0,
		},
		{
			name: "single audit",
			audits: []Audit{
				{
					URL:            "https://example.com",
					MissingTopics:  []string{"topic1", "topic2"},
					EntityCoverage: []EntityStatus{{Entity: "entity1", Status: "missing"}},
					UniqueAngles:   []string{"angle1"},
					QualityFlags:   []string{"thin_content"},
					ContentLength:  1000,
				},
			},
			wantLen: struct {
				gaps            int
				entitiesToCover int
				entitiesToImpr  int
				angles          int
				qualityIssues   int
			}{2, 1, 0, 1, 1},
			wantAvgWords: 1000.0,
		},
		{
			name: "multiple audits with aggregation",
			audits: []Audit{
				{
					URL:            "https://example1.com",
					MissingTopics:  []string{"topic1", "topic2"},
					EntityCoverage: []EntityStatus{{Entity: "entity1", Status: "missing"}, {Entity: "entity2", Status: "weak"}},
					UniqueAngles:   []string{"angle1", "angle2"},
					QualityFlags:   []string{"thin_content", "missing_h2"},
					ContentLength:  1000,
				},
				{
					URL:            "https://example2.com",
					MissingTopics:  []string{"topic1", "topic3"},
					EntityCoverage: []EntityStatus{{Entity: "entity1", Status: "missing"}, {Entity: "entity2", Status: "good"}},
					UniqueAngles:   []string{"angle2", "angle3"},
					QualityFlags:   []string{"thin_content"},
					ContentLength:  1500,
				},
				{
					URL:            "https://example3.com",
					MissingTopics:  []string{"topic2"},
					EntityCoverage: []EntityStatus{{Entity: "entity1", Status: "good"}, {Entity: "entity3", Status: "weak"}},
					UniqueAngles:   []string{"angle1"},
					QualityFlags:   []string{"missing_h2"},
					ContentLength:  2000,
				},
			},
			wantLen: struct {
				gaps            int
				entitiesToCover int
				entitiesToImpr  int
				angles          int
				qualityIssues   int
			}{3, 1, 2, 3, 2},
			wantAvgWords: 1500.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeAudits(tt.audits)

			if result == nil {
				t.Fatal("AnalyzeAudits returned nil")
			}

			if len(result.ContentGaps) != tt.wantLen.gaps {
				t.Errorf("ContentGaps length = %d, want %d", len(result.ContentGaps), tt.wantLen.gaps)
			}

			if len(result.EntitiesToCover) != tt.wantLen.entitiesToCover {
				t.Errorf("EntitiesToCover length = %d, want %d", len(result.EntitiesToCover), tt.wantLen.entitiesToCover)
			}

			if len(result.EntitiesToImprove) != tt.wantLen.entitiesToImpr {
				t.Errorf("EntitiesToImprove length = %d, want %d", len(result.EntitiesToImprove), tt.wantLen.entitiesToImpr)
			}

			if len(result.CompetitorAngles) != tt.wantLen.angles {
				t.Errorf("CompetitorAngles length = %d, want %d", len(result.CompetitorAngles), tt.wantLen.angles)
			}

			if len(result.TopQualityIssues) != tt.wantLen.qualityIssues {
				t.Errorf("TopQualityIssues length = %d, want %d", len(result.TopQualityIssues), tt.wantLen.qualityIssues)
			}

			if result.AverageWordCount != tt.wantAvgWords {
				t.Errorf("AverageWordCount = %f, want %f", result.AverageWordCount, tt.wantAvgWords)
			}
		})
	}
}

func TestAnalyzeAudits_GapSorting(t *testing.T) {
	audits := []Audit{{MissingTopics: []string{"topic1", "topic2", "topic3"}}, {MissingTopics: []string{"topic1", "topic2"}}, {MissingTopics: []string{"topic1"}}}

	result := AnalyzeAudits(audits)

	if len(result.ContentGaps) == 0 {
		t.Fatal("Expected content gaps but got none")
	}

	if result.ContentGaps[0].Topic != "topic1" {
		t.Errorf("Expected most frequent topic to be 'topic1', got '%s'", result.ContentGaps[0].Topic)
	}

	if result.ContentGaps[0].MissingCount != 3 {
		t.Errorf("Expected topic1 missing count to be 3, got %d", result.ContentGaps[0].MissingCount)
	}
}

func TestAnalyzeAudits_EntityCoverage(t *testing.T) {
	audits := []Audit{
		{
			EntityCoverage: []EntityStatus{{Entity: "entity_missing", Status: "missing"}, {Entity: "entity_weak", Status: "weak"}, {Entity: "entity_good", Status: "good"}},
		},
		{
			EntityCoverage: []EntityStatus{{Entity: "entity_missing", Status: "missing"}, {Entity: "entity_weak", Status: "weak"}, {Entity: "entity_good", Status: "good"}},
		},
		{
			EntityCoverage: []EntityStatus{{Entity: "entity_missing", Status: "missing"}, {Entity: "entity_weak", Status: "good"}, {Entity: "entity_good", Status: "good"}},
		},
	}

	result := AnalyzeAudits(audits)

	foundMissing := false
	for _, e := range result.EntitiesToCover {
		if e.Entity == "entity_missing" && e.Status == "missing" && e.Occurrences == 3 {
			foundMissing = true
			break
		}
	}
	if !foundMissing {
		t.Error("Expected entity_missing in EntitiesToCover with 3 occurrences")
	}

	foundWeak := false
	for _, e := range result.EntitiesToImprove {
		if e.Entity == "entity_weak" && e.Status == "weak" && e.Occurrences == 2 {
			foundWeak = true
			break
		}
	}
	if !foundWeak {
		t.Error("Expected entity_weak in EntitiesToImprove with 2 occurrences")
	}

	for _, e := range result.EntitiesToCover {
		if e.Entity == "entity_good" {
			t.Error("entity_good should not be in EntitiesToCover")
		}
	}
	for _, e := range result.EntitiesToImprove {
		if e.Entity == "entity_good" {
			t.Error("entity_good should not be in EntitiesToImprove")
		}
	}
}

func TestAnalyzeAudits_IgnoreEmpty(t *testing.T) {
	audits := []Audit{{
		MissingTopics:  []string{"", "valid_topic", "missing/unknown"},
		EntityCoverage: []EntityStatus{{Entity: "", Status: "missing"}},
		UniqueAngles:   []string{"", "valid_angle"},
		QualityFlags:   []string{"", "valid_flag"},
	}}

	result := AnalyzeAudits(audits)

	if len(result.ContentGaps) != 1 || result.ContentGaps[0].Topic != "valid_topic" {
		t.Error("Should ignore empty and 'missing/unknown' topics")
	}

	if len(result.EntitiesToCover) != 0 {
		t.Error("Should ignore empty entities")
	}

	angleFound := false
	for _, angle := range result.CompetitorAngles {
		if angle == "valid_angle" {
			angleFound = true
		}
		if angle == "" {
			t.Error("Should not include empty angles")
		}
	}
	if !angleFound {
		t.Error("Should include valid angle")
	}

	if len(result.TopQualityIssues) != 1 || result.TopQualityIssues[0].Issue != "valid_flag" {
		t.Error("Should ignore empty quality flags")
	}
}

func TestLimitHelpers(t *testing.T) {
	gaps := []Gap{{Topic: "a"}, {Topic: "b"}, {Topic: "c"}}
	if len(limitGaps(gaps, 2)) != 2 {
		t.Errorf("limitGaps length mismatch")
	}
	if len(limitGaps(gaps, 10)) != 3 {
		t.Errorf("limitGaps should return full slice when limit >= len")
	}

	entities := []EntityAnalysis{{Entity: "a"}, {Entity: "b"}, {Entity: "c"}}
	if len(limitEntities(entities, 2)) != 2 {
		t.Errorf("limitEntities length mismatch")
	}
	if len(limitEntities(entities, 5)) != 3 {
		t.Errorf("limitEntities should return full slice when limit >= len")
	}

	issues := []QualityIssue{{Issue: "a"}, {Issue: "b"}, {Issue: "c"}, {Issue: "d"}}
	if len(limitQuality(issues, 2)) != 2 {
		t.Errorf("limitQuality length mismatch")
	}
	if len(limitQuality(issues, 5)) != 4 {
		t.Errorf("limitQuality should return full slice when limit >= len")
	}
}

func TestAnalyzeAudits_Limits(t *testing.T) {
	audits := make([]Audit, 20)
	for i := 0; i < 20; i++ {
		audits[i] = Audit{
			MissingTopics: []string{
				"topic1", "topic2", "topic3", "topic4", "topic5",
				"topic6", "topic7", "topic8", "topic9", "topic10",
				"topic11", "topic12", "topic13", "topic14", "topic15",
				"topic16", "topic17", "topic18", "topic19", "topic20",
			},
			EntityCoverage: []EntityStatus{
				{Entity: "e1", Status: "missing"}, {Entity: "e2", Status: "missing"},
				{Entity: "e3", Status: "missing"}, {Entity: "e4", Status: "missing"},
				{Entity: "e5", Status: "missing"}, {Entity: "e6", Status: "missing"},
				{Entity: "e7", Status: "missing"}, {Entity: "e8", Status: "missing"},
				{Entity: "e9", Status: "missing"}, {Entity: "e10", Status: "missing"},
				{Entity: "e11", Status: "missing"}, {Entity: "e12", Status: "missing"},
				{Entity: "e13", Status: "missing"}, {Entity: "e14", Status: "missing"},
				{Entity: "e15", Status: "missing"}, {Entity: "e16", Status: "missing"},
				{Entity: "e17", Status: "missing"}, {Entity: "e18", Status: "missing"},
				{Entity: "e19", Status: "missing"}, {Entity: "e20", Status: "missing"},
				{Entity: "e21", Status: "missing"},
			},
			QualityFlags: []string{
				"issue1", "issue2", "issue3", "issue4", "issue5",
				"issue6", "issue7", "issue8", "issue9", "issue10",
				"issue11",
			},
		}
	}

	result := AnalyzeAudits(audits)

	if len(result.ContentGaps) > 15 {
		t.Errorf("ContentGaps should be limited to 15, got %d", len(result.ContentGaps))
	}

	if len(result.EntitiesToCover) > 20 {
		t.Errorf("EntitiesToCover should be limited to 20, got %d", len(result.EntitiesToCover))
	}

	if len(result.TopQualityIssues) > 10 {
		t.Errorf("TopQualityIssues should be limited to 10, got %d", len(result.TopQualityIssues))
	}
}
