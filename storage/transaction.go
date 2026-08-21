package storage

import "sitepreflight/model"

type Transaction struct {
	store   *Store
	records []model.Record
	audits  []model.AuditEvent
}

func (s *Store) Begin() *Transaction { return &Transaction{store: s} }
func (t *Transaction) AddRecord(r model.Record) *Transaction {
	t.records = append(t.records, r)
	return t
}
func (t *Transaction) AddAudit(a model.AuditEvent) *Transaction {
	t.audits = append(t.audits, a)
	return t
}
func (t *Transaction) Commit() error {
	for _, r := range t.records {
		if e := t.store.PutRecord(r); e != nil {
			return e
		}
	}
	for _, a := range t.audits {
		if e := t.store.PutAudit(a); e != nil {
			return e
		}
	}
	return nil
}
func (t *Transaction) Count() int { return len(t.records) + len(t.audits) }
