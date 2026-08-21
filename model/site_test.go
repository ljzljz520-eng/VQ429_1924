package model

import "testing"

func TestRecordValidation(t *testing.T) {
	if (Record{ID: "1", BatchID: "b", Name: "n"}).Validate() != nil {
		t.Fatal("valid")
	}
	if (Record{}).Validate() == nil {
		t.Fatal("invalid")
	}
}
