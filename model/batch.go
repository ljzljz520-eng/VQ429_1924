package model

import (
	"fmt"
	"sort"
)

type BatchSummary struct {
	BatchID     string
	Total       int
	Confirmed   int
	Archived    int
	NeedsReview int
	Owners      []string
	Domains     []string
	Ready       bool
}

func SummarizeBatch(batchID string, records []Record, policy Policy) BatchSummary {
	selected := BatchRecords(records, batchID)
	summary := BatchSummary{BatchID: batchID, Total: len(selected), Owners: Owners(selected), Domains: Domains(selected)}
	for _, record := range selected {
		switch record.Status {
		case StatusConfirmed:
			summary.Confirmed++
		case StatusArchived:
			summary.Archived++
		default:
			summary.NeedsReview++
		}
	}
	summary.Ready = summary.Total > 0 && summary.Confirmed+summary.Archived == summary.Total
	for _, record := range selected {
		if !Assess(record, policy).Ready() && !record.IsArchived() {
			summary.Ready = false
		}
	}
	return summary
}

func (s BatchSummary) CompletionPercent() int {
	if s.Total == 0 {
		return 0
	}
	return (s.Confirmed + s.Archived) * 100 / s.Total
}

func (s BatchSummary) Label() string {
	return fmt.Sprintf("%s: %d/%d complete", s.BatchID, s.Confirmed+s.Archived, s.Total)
}

func SortBatchSummaries(values []BatchSummary) []BatchSummary {
	out := append([]BatchSummary(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].BatchID < out[j].BatchID })
	return out
}

func SummarizeAll(records []Record, policy Policy) []BatchSummary {
	ids := make(map[string]bool)
	for _, record := range records {
		ids[record.BatchID] = true
	}
	summaries := make([]BatchSummary, 0, len(ids))
	for id := range ids {
		summaries = append(summaries, SummarizeBatch(id, records, policy))
	}
	return SortBatchSummaries(summaries)
}

func BatchIDs(records []Record) []string {
	ids := make([]string, 0)
	for _, record := range records {
		ids = append(ids, record.BatchID)
	}
	return UniqueStrings(ids)
}

func RecordIDs(records []Record) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	return ids
}

func ReplaceRecord(records []Record, replacement Record) []Record {
	out := make([]Record, 0, len(records))
	replaced := false
	for _, record := range records {
		if record.ID == replacement.ID {
			out = append(out, replacement)
			replaced = true
		} else {
			out = append(out, record)
		}
	}
	if !replaced {
		out = append(out, replacement)
	}
	return SortRecords(out)
}

func RemoveRecord(records []Record, id string) []Record {
	out := make([]Record, 0, len(records))
	for _, record := range records {
		if record.ID != id {
			out = append(out, record)
		}
	}
	return out
}
