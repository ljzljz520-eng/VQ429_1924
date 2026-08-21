package review

import (
	"fmt"
	"sitepreflight/model"
	"sitepreflight/storage"
)

type Decision struct {
	RecordID   string
	Reviewer   string
	Outcome    string
	Reason     string
	Assessment model.Assessment
}

func (s *Service) Assess(id string, policy model.Policy) (model.Assessment, error) {
	record, err := s.store.GetRecord(id)
	if err != nil {
		return model.Assessment{}, err
	}
	return model.Assess(record, policy), nil
}

func (s *Service) Decide(id, reviewer, outcome, reason string, policy model.Policy) (Decision, error) {
	record, err := s.store.GetRecord(id)
	if err != nil {
		return Decision{}, err
	}
	assessment := model.Assess(record, policy)
	if outcome == "confirm" && !assessment.Ready() {
		return Decision{}, fmt.Errorf("record %s has unresolved readiness issues", id)
	}
	if outcome != "confirm" && outcome != "hold" && outcome != "reject" {
		return Decision{}, fmt.Errorf("unsupported review outcome %s", outcome)
	}
	event := model.AuditEvent{ID: id + "-decision-" + reviewer, RecordID: id, Actor: reviewer, Action: "decision", Detail: outcome + ": " + reason}
	if err := s.store.PutAudit(event); err != nil {
		return Decision{}, err
	}
	return Decision{RecordID: id, Reviewer: reviewer, Outcome: outcome, Reason: reason, Assessment: assessment}, nil
}

func (s *Service) DecisionHistory(id string) ([]model.AuditEvent, error) {
	events, err := s.store.ListAudits(id)
	if err != nil {
		return nil, err
	}
	return model.FilterEvents(events, "decision"), nil
}

func (s *Service) RequiresSecondReview(record model.Record) bool {
	return record.Status == model.StatusInReview && len(record.Tags) > 3
}

func (s *Service) ReviewLabel(record model.Record) string {
	if record.IsArchived() {
		return "read-only archive"
	}
	if record.IsConfirmed() {
		return "ready to publish"
	}
	return "review required"
}

func (s *Service) ReviewSummary(id string, policy model.Policy) (string, error) {
	assessment, err := s.Assess(id, policy)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s: %s", id, assessment.Level), nil
}

var _ = storage.Store{}
