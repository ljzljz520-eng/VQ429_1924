package storage

import (
	"encoding/json"
	"sitepreflight/model"
)

type Snapshot struct {
	Records     []model.Record
	Audits      []model.AuditEvent
	Workflows   []model.Workflow
	Attachments []model.Attachment
}

func (s *Store) Snapshot() (Snapshot, error) {
	records, err := s.ListRecords()
	if err != nil {
		return Snapshot{}, err
	}
	audits, err := s.ListAudits("")
	if err != nil {
		return Snapshot{}, err
	}
	workflows, err := s.ListWorkflows()
	if err != nil {
		return Snapshot{}, err
	}
	attachments, err := s.ListAttachments("")
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Records: records, Audits: audits, Workflows: workflows, Attachments: attachments}, nil
}

func (s Snapshot) Marshal() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func DecodeSnapshot(data []byte) (Snapshot, error) {
	var snapshot Snapshot
	err := json.Unmarshal(data, &snapshot)
	return snapshot, err
}

func (s Snapshot) Counts() map[string]int {
	return map[string]int{"records": len(s.Records), "audits": len(s.Audits), "workflows": len(s.Workflows), "attachments": len(s.Attachments)}
}

func (s *Store) Restore(snapshot Snapshot) error {
	transaction := s.Begin()
	for _, record := range snapshot.Records {
		transaction.AddRecord(record)
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	for _, audit := range snapshot.Audits {
		if err := s.PutAudit(audit); err != nil {
			return err
		}
	}
	for _, workflow := range snapshot.Workflows {
		if err := s.PutWorkflow(workflow); err != nil {
			return err
		}
	}
	for _, attachment := range snapshot.Attachments {
		if err := s.PutAttachment(attachment); err != nil {
			return err
		}
	}
	return nil
}
