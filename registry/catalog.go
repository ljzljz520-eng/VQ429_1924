package registry

import (
	"fmt"
	"sitepreflight/model"
	"sitepreflight/storage"
)

type Catalog struct{ store *storage.Store }

func New(s *storage.Store) *Catalog { return &Catalog{store: s} }
func (c *Catalog) Register(r model.Record) error {
	if r.Status == "" {
		r.Status = model.StatusDraft
	}
	if e := r.Validate(); e != nil {
		return e
	}
	return c.store.PutRecord(r)
}
func (c *Catalog) Update(r model.Record) error {
	old, e := c.store.GetRecord(r.ID)
	if e != nil {
		return e
	}
	if r.BatchID == "" {
		r.BatchID = old.BatchID
	}
	if r.Status == "" {
		r.Status = old.Status
	}
	return c.store.PutRecord(r)
}
func (c *Catalog) Confirm(id, actor string) error {
	r, e := c.store.GetRecord(id)
	if e != nil {
		return e
	}
	if r.Status != model.StatusInReview {
		return fmt.Errorf("record is not in review")
	}
	r.Status = model.StatusConfirmed
	if e = c.store.PutRecord(r); e != nil {
		return e
	}
	return c.store.PutAudit(model.AuditEvent{ID: id + "-confirm", RecordID: id, Actor: actor, Action: "confirm", Detail: "migration preflight confirmed"})
}
func (c *Catalog) Archive(id, actor string) error {
	r, e := c.store.GetRecord(id)
	if e != nil {
		return e
	}
	if !r.IsConfirmed() {
		return fmt.Errorf("record must be confirmed")
	}
	r.Status = model.StatusArchived
	if e = c.store.PutRecord(r); e != nil {
		return e
	}
	return c.store.PutAudit(model.AuditEvent{ID: id + "-archive", RecordID: id, Actor: actor, Action: "archive", Detail: "retained for migration evidence"})
}
func (c *Catalog) SubmitForReview(id, actor string) error {
	r, e := c.store.GetRecord(id)
	if e != nil {
		return e
	}
	if r.Status != model.StatusDraft {
		return fmt.Errorf("only drafts can be submitted")
	}
	r.Status = model.StatusInReview
	if e = c.store.PutRecord(r); e != nil {
		return e
	}
	return c.store.PutAudit(model.AuditEvent{ID: id + "-review", RecordID: id, Actor: actor, Action: "submit", Detail: "review requested"})
}
func (c *Catalog) Attach(a model.Attachment) error { return c.store.PutAttachment(a) }
