package model

import (
	"fmt"
	"sort"
	"strings"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskModerate RiskLevel = "moderate"
	RiskHigh     RiskLevel = "high"
)

type Assessment struct {
	RecordID        string
	BatchID         string
	Score           int
	Level           RiskLevel
	MissingChecks   []string
	BlockingIssues  []string
	Recommendations []string
}

func Assess(record Record, policy Policy) Assessment {
	findings := policy.ValidateRecord(record)
	assessment := Assessment{RecordID: record.ID, BatchID: record.BatchID}
	for _, finding := range findings {
		switch finding.Severity {
		case SeverityBlocker:
			assessment.Score += 40
			assessment.BlockingIssues = append(assessment.BlockingIssues, finding.Message)
		case SeverityWarning:
			assessment.Score += 10
			assessment.MissingChecks = append(assessment.MissingChecks, finding.Message)
		case SeverityInfo:
			assessment.Score += 2
			assessment.Recommendations = append(assessment.Recommendations, finding.Message)
		}
	}
	if record.Status == StatusDraft {
		assessment.Score += 5
	}
	if record.Status == StatusArchived {
		assessment.Score = 0
	}
	assessment.Level = RiskForScore(assessment.Score)
	assessment.MissingChecks = UniqueStrings(assessment.MissingChecks)
	assessment.BlockingIssues = UniqueStrings(assessment.BlockingIssues)
	assessment.Recommendations = UniqueStrings(assessment.Recommendations)
	return assessment
}

func RiskForScore(score int) RiskLevel {
	switch {
	case score >= 40:
		return RiskHigh
	case score >= 10:
		return RiskModerate
	default:
		return RiskLow
	}
}

func (a Assessment) Ready() bool {
	return a.Level == RiskLow && len(a.BlockingIssues) == 0
}

func (a Assessment) String() string {
	return fmt.Sprintf("%s score=%d level=%s", a.RecordID, a.Score, a.Level)
}

func AssessAll(records []Record, policy Policy) []Assessment {
	assessments := make([]Assessment, 0, len(records))
	for _, record := range records {
		assessments = append(assessments, Assess(record, policy))
	}
	sort.Slice(assessments, func(i, j int) bool {
		if assessments[i].Score == assessments[j].Score {
			return assessments[i].RecordID < assessments[j].RecordID
		}
		return assessments[i].Score > assessments[j].Score
	})
	return assessments
}

func HighestRisk(records []Record, policy Policy) (Assessment, bool) {
	assessments := AssessAll(records, policy)
	if len(assessments) == 0 {
		return Assessment{}, false
	}
	return assessments[0], true
}

func ReadyForConfirmation(record Record, policy Policy) bool {
	return Assess(record, policy).Ready() && record.Status == StatusInReview
}

func AssessmentLines(assessment Assessment) []string {
	lines := []string{assessment.String()}
	for _, issue := range assessment.BlockingIssues {
		lines = append(lines, "blocker: "+issue)
	}
	for _, check := range assessment.MissingChecks {
		lines = append(lines, "check: "+check)
	}
	for _, recommendation := range assessment.Recommendations {
		lines = append(lines, "note: "+recommendation)
	}
	return lines
}

func MergeAssessments(values []Assessment) map[RiskLevel]int {
	result := map[RiskLevel]int{RiskLow: 0, RiskModerate: 0, RiskHigh: 0}
	for _, assessment := range values {
		result[assessment.Level]++
	}
	return result
}

func MissingChecklist(record Record, policy Policy) []string {
	missing := make([]string, 0)
	for _, item := range policy.RequiredChecklist {
		if !ContainsString(record.Checklist, item) {
			missing = append(missing, item)
		}
	}
	return missing
}

func RecommendationFor(record Record) []string {
	recommendations := make([]string, 0)
	if !record.HasTag("migration") {
		recommendations = append(recommendations, "add migration tag")
	}
	if record.Owner == "" {
		recommendations = append(recommendations, "assign an owner")
	}
	if strings.TrimSpace(record.Notes) == "" {
		recommendations = append(recommendations, "add handover note")
	}
	return recommendations
}
