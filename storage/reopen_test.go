package storage

import (
	"sitepreflight/model"
	"testing"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	p := t.TempDir() + "/persist.db"
	s, e := Open(p)
	if e != nil {
		t.Fatal(e)
	}
	r := model.Record{ID: "r", BatchID: "b", Name: "n"}
	if e = s.PutRecord(r); e != nil {
		t.Fatal(e)
	}
	if e = s.PutAudit(model.AuditEvent{ID: "a", RecordID: "r", Action: "create"}); e != nil {
		t.Fatal(e)
	}
	if e = s.PutWorkflow(model.Workflow{ID: "w", BatchID: "b", RecordIDs: []string{"r"}}); e != nil {
		t.Fatal(e)
	}
	if e = s.PutAttachment(model.Attachment{ID: "att", RecordID: "r", Filename: "x"}); e != nil {
		t.Fatal(e)
	}
	s.Close()
	s, e = Open(p)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	if _, e = s.GetRecord("r"); e != nil {
		t.Fatal(e)
	}
	if _, e = s.GetWorkflow("w"); e != nil {
		t.Fatal(e)
	}
	if len(mustAudits(s)) != 1 {
		t.Fatal("audit")
	}
	if len(mustAttachments(s)) != 1 {
		t.Fatal("attachment")
	}
}
func mustAudits(s *Store) []model.AuditEvent      { x, _ := s.ListAudits(""); return x }
func mustAttachments(s *Store) []model.Attachment { x, _ := s.ListAttachments(""); return x }
