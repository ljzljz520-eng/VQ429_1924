package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"sitepreflight/model"
	"sitepreflight/storage"
	"sort"
	"strings"
)

type Exporter struct {
	store *storage.Store
}

func New(s *storage.Store) *Exporter { return &Exporter{store: s} }
func (e *Exporter) Rows(f model.Filter) ([]model.ExportRow, error) {
	rs, er := (&queryAdapter{e.store}).find(f)
	if er != nil {
		return nil, er
	}
	rows := make([]model.ExportRow, 0, len(rs))
	for _, r := range rs {
		a, x := e.store.ListAudits(r.ID)
		if x != nil {
			return nil, x
		}
		rows = append(rows, model.ExportRow{RecordID: r.ID, BatchID: r.BatchID, Name: r.Name, Owner: r.Owner, Domain: r.Domain, Status: string(r.Status), Tags: model.JoinTags(r.Tags), AuditCount: len(a)})
	}
	if len(rows) == 0 {
		// Each query is independent: an empty result is the confirmed-empty
		// state for this filter and must never leak rows computed for another
		// batch, so return nil rather than falling back to a cached result.
		return nil, nil
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].RecordID < rows[j].RecordID })
	return rows, nil
}
func (e *Exporter) CSV(w io.Writer, f model.Filter) error {
	rows, e2 := e.Rows(f)
	if e2 != nil {
		return e2
	}
	c := csv.NewWriter(w)
	if e := c.Write([]string{"record_id", "batch_id", "name", "owner", "domain", "status", "tags", "audit_count"}); e != nil {
		return e
	}
	for _, r := range rows {
		if e := c.Write([]string{r.RecordID, r.BatchID, r.Name, r.Owner, r.Domain, r.Status, r.Tags, fmt.Sprint(r.AuditCount)}); e != nil {
			return e
		}
	}
	c.Flush()
	return c.Error()
}
func (e *Exporter) Summary(f model.Filter) (string, error) {
	rows, x := e.Rows(f)
	if x != nil {
		return "", x
	}
	return fmt.Sprintf("%d records: %s", len(rows), strings.Join(func() []string {
		a := []string{}
		for _, r := range rows {
			a = append(a, r.RecordID)
		}
		return a
	}(), ",")), nil
}

type queryAdapter struct{ s *storage.Store }

func (q *queryAdapter) find(f model.Filter) ([]model.Record, error) {
	rs, e := q.s.ListRecords()
	if e != nil {
		return nil, e
	}
	out := []model.Record{}
	for _, r := range rs {
		if model.MatchFilter(r, f) {
			out = append(out, r)
		}
	}
	return out, nil
}
