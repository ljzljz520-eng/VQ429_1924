package model

import (
	"fmt"
	"sort"
	"strings"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityBlocker Severity = "blocker"
)

type Finding struct {
	Code     string
	Severity Severity
	Field    string
	Message  string
}

type Policy struct {
	RequiredChecklist []string
	RequiredTags      []string
	AllowedStatuses   []SiteStatus
	MinimumDomainPart int
}

func DefaultPolicy() Policy {
	return Policy{
		RequiredChecklist: []string{"dns", "tls", "redirects", "rollback"},
		RequiredTags:      []string{"migration"},
		AllowedStatuses: []SiteStatus{
			StatusDraft,
			StatusInReview,
			StatusConfirmed,
			StatusArchived,
		},
		MinimumDomainPart: 2,
	}
}

func (p Policy) ValidateRecord(r Record) []Finding {
	findings := make([]Finding, 0, 8)
	if strings.TrimSpace(r.ID) == "" {
		findings = append(findings, Finding{"record.id", SeverityBlocker, "id", "record identifier is required"})
	}
	if strings.TrimSpace(r.BatchID) == "" {
		findings = append(findings, Finding{"record.batch", SeverityBlocker, "batch", "batch identifier is required"})
	}
	if strings.TrimSpace(r.Name) == "" {
		findings = append(findings, Finding{"record.name", SeverityBlocker, "name", "site name is required"})
	}
	if !p.AllowsStatus(r.Status) {
		findings = append(findings, Finding{"record.status", SeverityBlocker, "status", "unsupported lifecycle status"})
	}
	if !p.ValidDomain(r.Domain) {
		findings = append(findings, Finding{"record.domain", SeverityBlocker, "domain", "domain must have two labels"})
	}
	for _, item := range p.RequiredChecklist {
		if !ContainsString(r.Checklist, item) {
			findings = append(findings, Finding{"checklist." + item, SeverityWarning, "checklist", item + " has not been checked"})
		}
	}
	for _, tag := range p.RequiredTags {
		if !r.HasTag(tag) {
			findings = append(findings, Finding{"tag." + tag, SeverityInfo, "tags", tag + " tag is recommended"})
		}
	}
	return SortFindings(findings)
}

func (p Policy) AllowsStatus(status SiteStatus) bool {
	for _, allowed := range p.AllowedStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

func (p Policy) ValidDomain(domain string) bool {
	parts := strings.Split(strings.TrimSpace(domain), ".")
	if len(parts) < p.MinimumDomainPart {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
	}
	return true
}

func (p Policy) CanSubmit(r Record) error {
	if r.Status != StatusDraft {
		return fmt.Errorf("record %s is not a draft", r.ID)
	}
	return p.BlockerError(r)
}

func (p Policy) CanConfirm(r Record) error {
	if r.Status != StatusInReview {
		return fmt.Errorf("record %s is not under review", r.ID)
	}
	return p.BlockerError(r)
}

func (p Policy) CanArchive(r Record) error {
	if r.Status != StatusConfirmed {
		return fmt.Errorf("record %s is not confirmed", r.ID)
	}
	return nil
}

func (p Policy) BlockerError(r Record) error {
	for _, finding := range p.ValidateRecord(r) {
		if finding.Severity == SeverityBlocker {
			return fmt.Errorf("%s: %s", finding.Field, finding.Message)
		}
	}
	return nil
}

func ContainsString(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), expected) {
			return true
		}
	}
	return false
}

func SortFindings(values []Finding) []Finding {
	out := append([]Finding(nil), values...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Severity == out[j].Severity {
			return out[i].Code < out[j].Code
		}
		return severityRank(out[i].Severity) > severityRank(out[j].Severity)
	})
	return out
}

func severityRank(value Severity) int {
	switch value {
	case SeverityBlocker:
		return 3
	case SeverityWarning:
		return 2
	default:
		return 1
	}
}

func FindingSummary(findings []Finding) map[Severity]int {
	summary := map[Severity]int{
		SeverityInfo:    0,
		SeverityWarning: 0,
		SeverityBlocker: 0,
	}
	for _, finding := range findings {
		summary[finding.Severity]++
	}
	return summary
}

func ChecklistProgress(record Record, required []string) (int, int) {
	complete := 0
	for _, item := range required {
		if ContainsString(record.Checklist, item) {
			complete++
		}
	}
	return complete, len(required)
}

func AddChecklistItem(record Record, item string) Record {
	item = strings.TrimSpace(item)
	if item == "" || ContainsString(record.Checklist, item) {
		return record.Clone()
	}
	copy := record.Clone()
	copy.Checklist = append(copy.Checklist, item)
	return copy
}

func RemoveChecklistItem(record Record, item string) Record {
	copy := record.Clone()
	copy.Checklist = make([]string, 0, len(record.Checklist))
	for _, value := range record.Checklist {
		if !strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(item)) {
			copy.Checklist = append(copy.Checklist, value)
		}
	}
	return copy
}

func AddTag(record Record, tag string) Record {
	copy := record.Clone()
	copy.Tags = UniqueStrings(append(copy.Tags, tag))
	return copy
}

func RemoveTag(record Record, tag string) Record {
	copy := record.Clone()
	copy.Tags = make([]string, 0, len(record.Tags))
	for _, value := range record.Tags {
		if !strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(tag)) {
			copy.Tags = append(copy.Tags, value)
		}
	}
	return copy
}

func RecordLabel(record Record) string {
	return strings.TrimSpace(record.Name) + " (" + strings.TrimSpace(record.Domain) + ")"
}

func NormalizeRecord(record Record) Record {
	copy := record.Clone()
	copy.ID = strings.TrimSpace(copy.ID)
	copy.BatchID = strings.TrimSpace(copy.BatchID)
	copy.Name = strings.TrimSpace(copy.Name)
	copy.Owner = strings.TrimSpace(copy.Owner)
	copy.Domain = strings.ToLower(strings.TrimSpace(copy.Domain))
	copy.Notes = strings.TrimSpace(copy.Notes)
	copy.Tags = UniqueStrings(copy.Tags)
	copy.Checklist = UniqueStrings(copy.Checklist)
	return copy
}

func BatchRecords(records []Record, batchID string) []Record {
	out := make([]Record, 0)
	for _, record := range records {
		if record.BatchID == batchID {
			out = append(out, record)
		}
	}
	return SortRecords(out)
}

func StatusCounts(records []Record) map[SiteStatus]int {
	counts := map[SiteStatus]int{}
	for _, record := range records {
		counts[record.Status]++
	}
	return counts
}

func Owners(records []Record) []string {
	owners := make([]string, 0, len(records))
	for _, record := range records {
		owners = append(owners, record.Owner)
	}
	return UniqueStrings(owners)
}

func Domains(records []Record) []string {
	domains := make([]string, 0, len(records))
	for _, record := range records {
		domains = append(domains, record.Domain)
	}
	return UniqueStrings(domains)
}

func ConfirmedRecords(records []Record) []Record {
	out := make([]Record, 0)
	for _, record := range records {
		if record.IsConfirmed() {
			out = append(out, record)
		}
	}
	return SortRecords(out)
}

func ActiveRecords(records []Record) []Record {
	out := make([]Record, 0)
	for _, record := range records {
		if !record.IsArchived() {
			out = append(out, record)
		}
	}
	return SortRecords(out)
}
