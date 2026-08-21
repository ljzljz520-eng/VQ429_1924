package model

import "testing"

func TestFilterAndSort(t *testing.T) {
	r := Record{ID: "1", BatchID: "b", Name: "Site", Status: StatusConfirmed, Tags: []string{"Blue"}}
	if !MatchFilter(r, Filter{BatchID: "b", ConfirmedOnly: true, Tags: []string{"blue"}}) {
		t.Fatal("match")
	}
	if len(SortRecords([]Record{{ID: "b"}, {ID: "a"}})) != 2 {
		t.Fatal("sort")
	}
}
