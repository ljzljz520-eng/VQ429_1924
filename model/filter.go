package model

import "strings"

func NormalizeFilter(f Filter) Filter {
	f.BatchID = strings.TrimSpace(f.BatchID)
	f.Owner = strings.TrimSpace(f.Owner)
	f.Status = strings.ToLower(strings.TrimSpace(f.Status))
	f.Query = strings.ToLower(strings.TrimSpace(f.Query))
	return f
}
func MatchFilter(r Record, f Filter) bool {
	f = NormalizeFilter(f)
	if f.BatchID != "" && r.BatchID != f.BatchID {
		return false
	}
	if f.Owner != "" && !strings.EqualFold(r.Owner, f.Owner) {
		return false
	}
	if f.Status != "" && strings.ToLower(string(r.Status)) != f.Status {
		return false
	}
	if f.ConfirmedOnly && !r.IsConfirmed() {
		return false
	}
	if f.Query != "" && !strings.Contains(r.SearchText(), f.Query) {
		return false
	}
	for _, tag := range f.Tags {
		if !r.HasTag(tag) {
			return false
		}
	}
	return true
}
func SortRecords(rs []Record) []Record {
	out := append([]Record(nil), rs...)
	for i := 1; i < len(out); i++ {
		key := out[i]
		j := i - 1
		for j >= 0 && out[j].ID > key.ID {
			out[j+1] = out[j]
			j--
		}
		out[j+1] = key
	}
	return out
}
func UniqueStrings(xs []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		x = strings.TrimSpace(x)
		if x != "" && !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
func JoinTags(xs []string) string { return strings.Join(UniqueStrings(xs), ",") }
