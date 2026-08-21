package collaboration

import (
	"fmt"
	"sitepreflight/model"
	"sitepreflight/storage"
)

type Board struct{ store *storage.Store }

func New(s *storage.Store) *Board { return &Board{store: s} }
func (b *Board) AddComment(recordID, actor, text string) error {
	if text == "" {
		return fmt.Errorf("comment empty")
	}
	return b.store.PutAudit(model.AuditEvent{ID: recordID + "-comment-" + actor, RecordID: recordID, Actor: actor, Action: "comment", Detail: text})
}
func (b *Board) History(recordID string) ([]model.AuditEvent, error) {
	return b.store.ListAudits(recordID)
}
func (b *Board) Latest(recordID string) (model.AuditEvent, error) {
	xs, e := b.History(recordID)
	if e != nil {
		return model.AuditEvent{}, e
	}
	if len(xs) == 0 {
		return model.AuditEvent{}, fmt.Errorf("no history")
	}
	return xs[len(xs)-1], nil
}
func (b *Board) Assign(recordID, owner string) error {
	r, e := b.store.GetRecord(recordID)
	if e != nil {
		return e
	}
	r.Owner = owner
	if e = b.store.PutRecord(r); e != nil {
		return e
	}
	return b.AddComment(recordID, owner, "assigned owner")
}
