package storage

import (
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
	"path/filepath"
	"sitepreflight/model"
	"sync"
)

var bucketRecords = []byte("records")
var bucketAudits = []byte("audits")
var bucketWorkflows = []byte("workflows")
var bucketAttachments = []byte("attachments")

type Store struct {
	db   *bbolt.DB
	mu   sync.RWMutex
	path string
}

func Open(path string) (*Store, error) {
	db, err := bbolt.Open(filepath.Clean(path), 0600, nil)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, b := range [][]byte{bucketRecords, bucketAudits, bucketWorkflows, bucketAttachments} {
			if _, e := tx.CreateBucketIfNotExists(b); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}
func (s *Store) Path() string { return s.path }
func (s *Store) PutRecord(r model.Record) error {
	if err := r.Validate(); err != nil {
		return err
	}
	b, e := r.Marshal()
	if e != nil {
		return e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketRecords).Put([]byte(r.ID), b) })
}
func (s *Store) GetRecord(id string) (model.Record, error) {
	var r model.Record
	s.mu.RLock()
	defer s.mu.RUnlock()
	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket(bucketRecords).Get([]byte(id))
		if v == nil {
			return fmt.Errorf("record %s not found", id)
		}
		var e error
		r, e = model.DecodeRecord(v)
		return e
	})
	return r, err
}
func (s *Store) DeleteRecord(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketRecords).Delete([]byte(id)) })
}
func (s *Store) ListRecords() ([]model.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []model.Record{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketRecords).ForEach(func(_, v []byte) error {
			r, e := model.DecodeRecord(v)
			if e != nil {
				return e
			}
			out = append(out, r)
			return nil
		})
	})
	return model.SortRecords(out), err
}
func (s *Store) PutAudit(a model.AuditEvent) error {
	if a.At == 0 {
		a.At = 1
	}
	b, e := a.Marshal()
	if e != nil {
		return e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketAudits).Put([]byte(a.ID), b) })
}
func (s *Store) ListAudits(recordID string) ([]model.AuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []model.AuditEvent{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAudits).ForEach(func(_, v []byte) error {
			a, e := model.DecodeAudit(v)
			if e != nil {
				return e
			}
			if recordID == "" || a.RecordID == recordID {
				out = append(out, a)
			}
			return nil
		})
	})
	return out, err
}
func (s *Store) PutWorkflow(w model.Workflow) error {
	b, e := w.Marshal()
	if e != nil {
		return e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketWorkflows).Put([]byte(w.ID), b) })
}
func (s *Store) GetWorkflow(id string) (model.Workflow, error) {
	var w model.Workflow
	s.mu.RLock()
	defer s.mu.RUnlock()
	e := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket(bucketWorkflows).Get([]byte(id))
		if v == nil {
			return fmt.Errorf("workflow not found")
		}
		var x error
		w, x = model.DecodeWorkflow(v)
		return x
	})
	return w, e
}
func (s *Store) ListWorkflows() ([]model.Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []model.Workflow{}
	e := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketWorkflows).ForEach(func(_, v []byte) error {
			w, x := model.DecodeWorkflow(v)
			if x == nil {
				out = append(out, w)
			}
			return x
		})
	})
	return out, e
}
func (s *Store) PutAttachment(a model.Attachment) error {
	b, e := a.Marshal()
	if e != nil {
		return e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketAttachments).Put([]byte(a.ID), b) })
}
func (s *Store) ListAttachments(recordID string) ([]model.Attachment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []model.Attachment{}
	e := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAttachments).ForEach(func(_, v []byte) error {
			a, x := model.DecodeAttachment(v)
			if x != nil {
				return x
			}
			if recordID == "" || a.RecordID == recordID {
				out = append(out, a)
			}
			return nil
		})
	})
	return out, e
}
func (s *Store) ExportJSON() ([]byte, error) {
	rs, e := s.ListRecords()
	if e != nil {
		return nil, e
	}
	return json.MarshalIndent(rs, "", "  ")
}
