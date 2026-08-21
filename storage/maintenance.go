package storage

import (
	"fmt"
	"sitepreflight/model"
)

type Health struct {
	Records     int
	Audits      int
	Workflows   int
	Attachments int
}

func (s *Store) Health() (Health, error) {
	records, err := s.ListRecords()
	if err != nil {
		return Health{}, err
	}
	audits, err := s.ListAudits("")
	if err != nil {
		return Health{}, err
	}
	workflows, err := s.ListWorkflows()
	if err != nil {
		return Health{}, err
	}
	attachments, err := s.ListAttachments("")
	if err != nil {
		return Health{}, err
	}
	return Health{Records: len(records), Audits: len(audits), Workflows: len(workflows), Attachments: len(attachments)}, nil
}

func (h Health) Empty() bool {
	return h.Records == 0 && h.Audits == 0 && h.Workflows == 0 && h.Attachments == 0
}

func (h Health) Total() int {
	return h.Records + h.Audits + h.Workflows + h.Attachments
}

func (s *Store) RecordExists(id string) bool {
	_, err := s.GetRecord(id)
	return err == nil
}

func (s *Store) WorkflowExists(id string) bool {
	_, err := s.GetWorkflow(id)
	return err == nil
}

func (s *Store) RequireRecord(id string) (model.Record, error) {
	record, err := s.GetRecord(id)
	if err != nil {
		return model.Record{}, fmt.Errorf("required record %s: %w", id, err)
	}
	return record, nil
}

func (s *Store) RequireWorkflow(id string) (model.Workflow, error) {
	workflow, err := s.GetWorkflow(id)
	if err != nil {
		return model.Workflow{}, fmt.Errorf("required workflow %s: %w", id, err)
	}
	return workflow, nil
}

func (s *Store) AuditCount(recordID string) (int, error) {
	events, err := s.ListAudits(recordID)
	return len(events), err
}

func (s *Store) AttachmentCount(recordID string) (int, error) {
	attachments, err := s.ListAttachments(recordID)
	return len(attachments), err
}
