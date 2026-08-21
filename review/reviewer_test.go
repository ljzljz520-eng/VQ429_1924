package review

import (
	"sitepreflight/model"
	"sitepreflight/storage"
	"testing"
)

func TestReviewer(t *testing.T) {
	s, _ := storage.Open(t.TempDir() + "/r.db")
	defer s.Close()
	s.PutRecord(model.Record{ID: "r", BatchID: "b", Name: "n", Domain: "d", Checklist: []string{"dns", "tls"}})
	r := New(s)
	if e := r.ValidateReady("r"); e != nil {
		t.Fatal(e)
	}
	if len(mustChecklist(r, "r")) != 2 {
		t.Fatal("checklist")
	}
}
func mustChecklist(r *Service, id string) []string { x, _ := r.Checklist(id); return x }
