package workflow

import (
	"fmt"
	"sitepreflight/model"
	"sitepreflight/storage"
)

type Engine struct{ store *storage.Store }

func New(s *storage.Store) *Engine { return &Engine{store: s} }
func (e *Engine) Import(batch string, records []model.Record) (model.Workflow, error) {
	if batch == "" {
		return model.Workflow{}, fmt.Errorf("batch required")
	}
	if err := ValidateBatch(batch, records); err != nil {
		return model.Workflow{}, err
	}
	ids := make([]string, 0, len(records))
	tx := e.store.Begin()
	for _, r := range records {
		ids = append(ids, r.ID)
		tx.AddRecord(r)
	}
	w := model.Workflow{ID: "wf-" + batch, BatchID: batch, Name: "migration preflight " + batch, State: "imported", RecordIDs: ids}
	if err := tx.Commit(); err != nil {
		return w, err
	}
	if err := e.store.PutWorkflow(w); err != nil {
		return w, err
	}
	return w, nil
}
func ValidateBatch(batch string, records []model.Record) error {
	if len(records) == 0 {
		return fmt.Errorf("batch has no records")
	}
	seen := map[string]bool{}
	for i, r := range records {
		if r.BatchID != batch {
			return fmt.Errorf("row %d has wrong batch", i)
		}
		if err := r.Validate(); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if seen[r.ID] {
			return fmt.Errorf("duplicate record %s", r.ID)
		}
		seen[r.ID] = true
	}
	return nil
}
func (e *Engine) Advance(id, state string) error {
	w, err := e.store.GetWorkflow(id)
	if err != nil {
		return err
	}
	if state != "imported" && state != "review" && state != "published" {
		return fmt.Errorf("invalid workflow state")
	}
	w.State = state
	return e.store.PutWorkflow(w)
}
func (e *Engine) Records(id string) ([]model.Record, error) {
	w, err := e.store.GetWorkflow(id)
	if err != nil {
		return nil, err
	}
	out := []model.Record{}
	for _, rid := range w.RecordIDs {
		r, x := e.store.GetRecord(rid)
		if x != nil {
			return nil, x
		}
		out = append(out, r)
	}
	return out, nil
}
func (e *Engine) Publish(id string) error {
	rs, err := e.Records(id)
	if err != nil {
		return err
	}
	for _, r := range rs {
		if !r.IsConfirmed() {
			return fmt.Errorf("record %s is not confirmed", r.ID)
		}
	}
	return e.Advance(id, "published")
}
