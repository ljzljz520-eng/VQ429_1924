package model

import (
	"fmt"
	"strings"
)

func StatusLabel(status SiteStatus) string {
	switch status {
	case StatusInReview:
		return "In review"
	case StatusConfirmed:
		return "Confirmed"
	case StatusArchived:
		return "Archived"
	default:
		return "Draft"
	}
}

func RiskLabel(level RiskLevel) string {
	switch level {
	case RiskHigh:
		return "High risk"
	case RiskModerate:
		return "Moderate risk"
	default:
		return "Low risk"
	}
}

func RecordSummary(record Record) string {
	return fmt.Sprintf("%s | %s | %s | %s", record.ID, StatusLabel(record.Status), record.Name, record.Domain)
}

func TagsLabel(record Record) string {
	if len(record.Tags) == 0 {
		return "none"
	}
	return JoinTags(record.Tags)
}

func ChecklistLabel(record Record, policy Policy) string {
	complete, total := ChecklistProgress(record, policy.RequiredChecklist)
	return fmt.Sprintf("%d/%d checks", complete, total)
}

func FindingLabels(findings []Finding) []string {
	labels := make([]string, 0, len(findings))
	for _, finding := range findings {
		labels = append(labels, strings.ToUpper(string(finding.Severity))+": "+finding.Message)
	}
	return labels
}

func JoinFindings(findings []Finding) string {
	return strings.Join(FindingLabels(findings), "; ")
}

func AssessmentSummary(record Record, policy Policy) string {
	assessment := Assess(record, policy)
	return fmt.Sprintf("%s: %s (%d)", record.ID, RiskLabel(assessment.Level), assessment.Score)
}

func BatchSummaryLines(summary BatchSummary) []string {
	return []string{
		summary.Label(),
		fmt.Sprintf("completion: %d%%", summary.CompletionPercent()),
		fmt.Sprintf("owners: %s", strings.Join(summary.Owners, ",")),
		fmt.Sprintf("domains: %s", strings.Join(summary.Domains, ",")),
	}
}

func PlanStepLabel(stepID, state string) string {
	return fmt.Sprintf("%s [%s]", stepID, state)
}

func NormalizeStatus(value string) SiteStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "review", "in_review", "in-review":
		return StatusInReview
	case "confirmed", "confirm":
		return StatusConfirmed
	case "archived", "archive":
		return StatusArchived
	default:
		return StatusDraft
	}
}
