package search

import (
	"sitepreflight/model"
	"sitepreflight/storage"
)

type Query struct{ store *storage.Store }

func New(s *storage.Store) *Query { return &Query{store: s} }
func (q *Query) Find(f model.Filter) ([]model.Record, error) {
	rs, e := q.store.ListRecords()
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
func (q *Query) FindConfirmed(batch string) ([]model.Record, error) {
	return q.Find(model.Filter{BatchID: batch, ConfirmedOnly: true})
}
func (q *Query) Count(f model.Filter) (int, error) { rs, e := q.Find(f); return len(rs), e }
func (q *Query) ByOwner(owner string) ([]model.Record, error) {
	return q.Find(model.Filter{Owner: owner})
}
func (q *Query) Batches() ([]string, error) {
	rs, e := q.store.ListRecords()
	if e != nil {
		return nil, e
	}
	seen := map[string]bool{}
	out := []string{}
	for _, r := range rs {
		if !seen[r.BatchID] {
			seen[r.BatchID] = true
			out = append(out, r.BatchID)
		}
	}
	return out, nil
}
