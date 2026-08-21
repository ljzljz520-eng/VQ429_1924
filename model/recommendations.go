package model

import "strings"

type Recommendation struct {
	Code     string
	Priority int
	Text     string
}

func Recommendations(record Record, policy Policy) []Recommendation {
	items := make([]Recommendation, 0)
	for _, check := range MissingChecklist(record, policy) {
		items = append(items, Recommendation{Code: "check-" + check, Priority: 2, Text: "complete " + check + " validation"})
	}
	if strings.TrimSpace(record.Owner) == "" {
		items = append(items, Recommendation{Code: "owner", Priority: 3, Text: "assign an accountable owner"})
	}
	if !record.HasTag("migration") {
		items = append(items, Recommendation{Code: "tag", Priority: 1, Text: "add the migration tag"})
	}
	return items
}

func HighestPriority(items []Recommendation) (Recommendation, bool) {
	if len(items) == 0 {
		return Recommendation{}, false
	}
	highest := items[0]
	for _, item := range items[1:] {
		if item.Priority > highest.Priority {
			highest = item
		}
	}
	return highest, true
}

func RecommendationTexts(items []Recommendation) []string {
	texts := make([]string, 0, len(items))
	for _, item := range items {
		texts = append(texts, item.Text)
	}
	return texts
}
