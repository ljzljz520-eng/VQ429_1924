package collaboration

import (
	"fmt"
	"sitepreflight/model"
)

type Role string

const (
	RoleOperator Role = "operator"
	RoleReviewer Role = "reviewer"
	RoleOwner    Role = "owner"
)

type Participant struct {
	Name string
	Role Role
}

func ValidRole(role Role) bool {
	return role == RoleOperator || role == RoleReviewer || role == RoleOwner
}

func (b *Board) AddParticipant(recordID string, participant Participant) error {
	if participant.Name == "" || !ValidRole(participant.Role) {
		return fmt.Errorf("invalid participant")
	}
	return b.AddComment(recordID, participant.Name, "joined as "+string(participant.Role))
}

func (b *Board) Participants(recordID string) ([]Participant, error) {
	events, err := b.History(recordID)
	if err != nil {
		return nil, err
	}
	participants := make([]Participant, 0)
	seen := map[string]bool{}
	for _, event := range events {
		if event.Actor != "" && !seen[event.Actor] {
			seen[event.Actor] = true
			participants = append(participants, Participant{Name: event.Actor, Role: RoleOperator})
		}
	}
	return participants, nil
}

func (b *Board) RecordDigest(recordID string) (string, error) {
	record, err := b.store.GetRecord(recordID)
	if err != nil {
		return "", err
	}
	events, err := b.History(recordID)
	if err != nil {
		return "", err
	}
	return record.ID + ":" + model.LastAction(events), nil
}

func (b *Board) CommentCount(recordID string) (int, error) {
	events, err := b.History(recordID)
	if err != nil {
		return 0, err
	}
	return model.EventCount(events, "comment"), nil
}
