package review

import (
	"fmt"
	"sitepreflight/model"
	"sitepreflight/storage"
)

type Service struct{ store *storage.Store }

func New(s *storage.Store) *Service { return &Service{store: s} }
func (s *Service) Inspect(id string) (model.Record, []model.AuditEvent, error) {
	r, e := s.store.GetRecord(id)
	if e != nil {
		return r, nil, e
	}
	a, e := s.store.ListAudits(id)
	return r, a, e
}
func (s *Service) Checklist(id string) ([]string, error) {
	r, e := s.store.GetRecord(id)
	if e != nil {
		return nil, e
	}
	out := append([]string(nil), r.Checklist...)
	if len(out) == 0 {
		return []string{"dns", "tls", "redirects", "rollback"}, nil
	}
	return out, nil
}
func (s *Service) ValidateReady(id string) error {
	r, e := s.store.GetRecord(id)
	if e != nil {
		return e
	}
	if r.Name == "" || r.Domain == "" {
		return fmt.Errorf("identity incomplete")
	}
	if len(r.Checklist) < 2 {
		return fmt.Errorf("checklist incomplete")
	}
	return nil
}
func (s *Service) AddNote(id, note string) error {
	r, e := s.store.GetRecord(id)
	if e != nil {
		return e
	}
	r.Notes = note
	return s.store.PutRecord(r)
}
