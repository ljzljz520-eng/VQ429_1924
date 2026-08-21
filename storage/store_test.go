package storage

import (
	"sitepreflight/model"
	"testing"
)

func TestStoreRoundTrip(t *testing.T) {
	s, e := Open(t.TempDir() + "/x.db")
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	r := model.Record{ID: "r", BatchID: "b", Name: "n"}
	if e = s.PutRecord(r); e != nil {
		t.Fatal(e)
	}
	got, e := s.GetRecord("r")
	if e != nil || got.Name != "n" {
		t.Fatal(e)
	}
}
