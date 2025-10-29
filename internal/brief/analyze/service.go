package analyze

import "sort"

// AnalyzeAudits agreguje dane z audytów i zwraca wynik analizy.
func AnalyzeAudits(audits []Audit) *Analysis {
	gapCount := make(map[string]int)
	for _, audit := range audits {
		for _, topic := range audit.MissingTopics {
			if topic != "" && topic != "missing/unknown" {
				gapCount[topic]++
			}
		}
	}

	gaps := make([]Gap, 0, len(gapCount))
	for topic, count := range gapCount {
		gaps = append(gaps, Gap{Topic: topic, MissingCount: count})
	}
	sort.Slice(gaps, func(i, j int) bool {
		return gaps[i].MissingCount > gaps[j].MissingCount
	})

	entityMap := make(map[string]map[string]int)
	for _, audit := range audits {
		for _, ec := range audit.EntityCoverage {
			if ec.Entity == "" {
				continue
			}
			if entityMap[ec.Entity] == nil {
				entityMap[ec.Entity] = make(map[string]int)
			}
			entityMap[ec.Entity][ec.Status]++
		}
	}

	var (
		entitiesToCover   []EntityAnalysis
		entitiesToImprove []EntityAnalysis
	)

	for entity, statuses := range entityMap {
		missingCount := statuses["missing"]
		weakCount := statuses["weak"]
		goodCount := statuses["good"]

		if missingCount > weakCount && missingCount > goodCount {
			entitiesToCover = append(entitiesToCover, EntityAnalysis{
				Entity:      entity,
				Status:      "missing",
				Occurrences: missingCount,
			})
		} else if weakCount > 0 && weakCount >= goodCount {
			entitiesToImprove = append(entitiesToImprove, EntityAnalysis{
				Entity:      entity,
				Status:      "weak",
				Occurrences: weakCount,
			})
		}
	}

	sort.Slice(entitiesToCover, func(i, j int) bool {
		return entitiesToCover[i].Occurrences > entitiesToCover[j].Occurrences
	})
	sort.Slice(entitiesToImprove, func(i, j int) bool {
		return entitiesToImprove[i].Occurrences > entitiesToImprove[j].Occurrences
	})

	anglesMap := make(map[string]bool)
	for _, audit := range audits {
		for _, angle := range audit.UniqueAngles {
			if angle != "" {
				anglesMap[angle] = true
			}
		}
	}
	angles := make([]string, 0, len(anglesMap))
	for angle := range anglesMap {
		angles = append(angles, angle)
	}

	totalWords := 0
	for _, audit := range audits {
		totalWords += audit.ContentLength
	}
	avgWords := 0.0
	if len(audits) > 0 {
		avgWords = float64(totalWords) / float64(len(audits))
	}

	qualityMap := make(map[string]int)
	for _, audit := range audits {
		for _, flag := range audit.QualityFlags {
			if flag != "" {
				qualityMap[flag]++
			}
		}
	}

	qualityIssues := make([]QualityIssue, 0, len(qualityMap))
	for issue, count := range qualityMap {
		qualityIssues = append(qualityIssues, QualityIssue{Issue: issue, Count: count})
	}
	sort.Slice(qualityIssues, func(i, j int) bool {
		return qualityIssues[i].Count > qualityIssues[j].Count
	})

	return &Analysis{
		ContentGaps:       limitGaps(gaps, 15),
		EntitiesToCover:   limitEntities(entitiesToCover, 20),
		EntitiesToImprove: limitEntities(entitiesToImprove, 15),
		CompetitorAngles:  angles,
		AverageWordCount:  avgWords,
		TopQualityIssues:  limitQuality(qualityIssues, 10),
	}
}

func limitGaps(s []Gap, limit int) []Gap {
	if len(s) < limit {
		return s
	}
	return s[:limit]
}

func limitEntities(s []EntityAnalysis, limit int) []EntityAnalysis {
	if len(s) < limit {
		return s
	}
	return s[:limit]
}

func limitQuality(s []QualityIssue, limit int) []QualityIssue {
	if len(s) < limit {
		return s
	}
	return s[:limit]
}
