package registry

import (
	"sitepreflight/model"
	"sitepreflight/storage"
	"testing"
)

func TestRegistryTransitions(t *testing.T) {
	s, _ := storage.Open(t.TempDir() + "/r.db")
	defer s.Close()
	c := New(s)
	if e := c.Register(model.Record{ID: "r", BatchID: "b", Name: "n", Domain: "d"}); e != nil {
		t.Fatal(e)
	}
	if e := c.SubmitForReview("r", "op"); e != nil {
		t.Fatal(e)
	}
	if e := c.Confirm("r", "rev"); e != nil {
		t.Fatal(e)
	}
	if e := c.Archive("r", "op"); e != nil {
		t.Fatal(e)
	}
}
