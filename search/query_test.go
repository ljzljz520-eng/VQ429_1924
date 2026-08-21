package search

import (
	"sitepreflight/model"
	"sitepreflight/storage"
	"testing"
)

func TestQuery(t *testing.T) {
	s, _ := storage.Open(t.TempDir() + "/q.db")
	defer s.Close()
	s.PutRecord(model.Record{ID: "r", BatchID: "b", Name: "n", Status: model.StatusConfirmed})
	q := New(s)
	rs, e := q.FindConfirmed("b")
	if e != nil || len(rs) != 1 {
		t.Fatal(e)
	}
}
