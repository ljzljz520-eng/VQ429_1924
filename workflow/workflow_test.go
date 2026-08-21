package workflow

import (
	"bytes"
	"sitepreflight/model"
	"sitepreflight/registry"
	"sitepreflight/report"
	"sitepreflight/search"
	"sitepreflight/storage"
	"testing"
)

func TestWorkflowImportReport(t *testing.T) {
	s, _ := storage.Open(t.TempDir() + "/w.db")
	defer s.Close()
	e := New(s)
	w, x := e.Import("b", []model.Record{{ID: "r", BatchID: "b", Name: "n", Status: model.StatusConfirmed}})
	if x != nil {
		t.Fatal(x)
	}
	if x = e.Publish(w.ID); x != nil {
		t.Fatal(x)
	}
}
func TestWorkflowCreateReviewArchive(t *testing.T) {
	s, err := storage.Open(t.TempDir() + "/w.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	catalog := registry.New(s)
	record := model.Record{ID: "site-1", BatchID: "429-21", Name: "North site", Owner: "ops", Domain: "north.example", Checklist: []string{"dns", "tls", "redirects", "rollback"}, Tags: []string{"migration"}}
	if err := catalog.Register(record); err != nil {
		t.Fatal(err)
	}
	if err := catalog.SubmitForReview(record.ID, "ops"); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Confirm(record.ID, "reviewer"); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Archive(record.ID, "ops"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetRecord(record.ID)
	if err != nil || got.Status != model.StatusArchived {
		t.Fatalf("archive status: %v", err)
	}
}
func TestWorkflowSearchUpdatePublish(t *testing.T) {
	s, err := storage.Open(t.TempDir() + "/w.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.PutRecord(model.Record{ID: "confirmed", BatchID: "429-21", Name: "Confirmed", Domain: "confirmed.example", Status: model.StatusConfirmed}); err != nil {
		t.Fatal(err)
	}
	if err := s.PutRecord(model.Record{ID: "draft", BatchID: "429-21", Name: "Draft", Domain: "draft.example", Status: model.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	results, err := search.New(s).FindConfirmed("429-21")
	if err != nil || len(results) != 1 || results[0].ID != "confirmed" {
		t.Fatalf("search results: %v", err)
	}
	updated := results[0]
	updated.Notes = "handover ready"
	if err := s.PutRecord(updated); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := report.New(s).CSV(&output, model.Filter{BatchID: "429-21", ConfirmedOnly: true}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte("confirmed")) {
		t.Fatal("export missing record")
	}
}
func TestBusiness21Regression(t *testing.T) {
	s, _ := storage.Open(t.TempDir() + "/w.db")
	defer s.Close()
	x := model.Record{ID: "r", BatchID: "429-21", Name: "n", Status: model.StatusConfirmed}
	s.PutRecord(x)
	ex := NewExport(s)
	if _, e := ex.Rows(model.Filter{BatchID: "429-21", ConfirmedOnly: true}); e != nil {
		t.Fatal(e)
	}
	rs, e := ex.Rows(model.Filter{BatchID: "other", ConfirmedOnly: true})
	if e != nil {
		t.Fatal(e)
	}
	if rs != nil {
		t.Fatalf("expected independent empty result, got %d", len(rs))
	}
}

type exporter interface {
	Rows(model.Filter) ([]model.ExportRow, error)
}

func NewExport(s *storage.Store) exporter { return report.New(s) }
