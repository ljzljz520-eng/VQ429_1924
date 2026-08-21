package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SiteStatus string

const (
	StatusDraft     SiteStatus = "draft"
	StatusInReview  SiteStatus = "in_review"
	StatusConfirmed SiteStatus = "confirmed"
	StatusArchived  SiteStatus = "archived"
)

type Record struct {
	ID, BatchID, Name, Owner, Domain, Notes string
	Status                                  SiteStatus
	Tags                                    []string
	Checklist                               []string
	CreatedAt, UpdatedAt                    int64
}
type AuditEvent struct {
	ID, RecordID, Actor, Action, Detail string
	At                                  int64
}
type Workflow struct {
	ID, BatchID, Name, State string
	RecordIDs                []string
	CreatedAt, UpdatedAt     int64
}
type Attachment struct {
	ID, RecordID, Filename, MediaType, ContentHash string
	Size                                           int64
}
type Filter struct {
	BatchID, Owner, Status, Query string
	Tags                          []string
	ConfirmedOnly                 bool
}
type ExportRow struct {
	RecordID, BatchID, Name, Owner, Domain, Status string
	Tags                                           string
	AuditCount                                     int
}

func (r Record) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("record id required")
	}
	if strings.TrimSpace(r.BatchID) == "" {
		return fmt.Errorf("batch id required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name required")
	}
	if r.Status == "" {
		r.Status = StatusDraft
	}
	return nil
}
func (r Record) IsConfirmed() bool { return r.Status == StatusConfirmed }
func (r Record) IsArchived() bool  { return r.Status == StatusArchived }
func (r Record) Clone() Record {
	r.Tags = append([]string(nil), r.Tags...)
	r.Checklist = append([]string(nil), r.Checklist...)
	return r
}
func (r Record) HasTag(tag string) bool {
	for _, t := range r.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}
func (r Record) SearchText() string {
	return strings.ToLower(strings.Join([]string{r.ID, r.BatchID, r.Name, r.Owner, r.Domain, r.Notes}, " "))
}
func (r Record) Marshal() ([]byte, error) { return json.Marshal(r) }
func DecodeRecord(b []byte) (Record, error) {
	var r Record
	err := json.Unmarshal(b, &r)
	return r, err
}
func (e AuditEvent) Marshal() ([]byte, error) { return json.Marshal(e) }
func DecodeAudit(b []byte) (AuditEvent, error) {
	var e AuditEvent
	err := json.Unmarshal(b, &e)
	return e, err
}
func (w Workflow) Marshal() ([]byte, error) { return json.Marshal(w) }
func DecodeWorkflow(b []byte) (Workflow, error) {
	var w Workflow
	err := json.Unmarshal(b, &w)
	return w, err
}
func (a Attachment) Marshal() ([]byte, error) { return json.Marshal(a) }
func DecodeAttachment(b []byte) (Attachment, error) {
	var a Attachment
	err := json.Unmarshal(b, &a)
	return a, err
}
