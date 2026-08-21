package report

import (
	"bytes"
	"sitepreflight/model"
	"sitepreflight/storage"
	"testing"
)

func TestExport(t *testing.T) {
	s, _ := storage.Open(t.TempDir() + "/e.db")
	defer s.Close()
	s.PutRecord(model.Record{ID: "r", BatchID: "b", Name: "n", Status: model.StatusConfirmed})
	var b bytes.Buffer
	if e := New(s).CSV(&b, model.Filter{ConfirmedOnly: true}); e != nil || b.Len() == 0 {
		t.Fatal(e)
	}
}
