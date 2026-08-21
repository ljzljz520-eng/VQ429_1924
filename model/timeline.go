package model

import (
	"sort"
	"strings"
)

type TimelineItem struct {
	RecordID string
	Status   SiteStatus
	Action   string
	Actor    string
	Detail   string
	At       int64
}

func BuildTimeline(record Record, events []AuditEvent) []TimelineItem {
	items := make([]TimelineItem, 0, len(events)+1)
	items = append(items, TimelineItem{RecordID: record.ID, Status: StatusDraft, Action: "created", Detail: "record registered"})
	for _, event := range events {
		status := statusForAction(event.Action)
		items = append(items, TimelineItem{RecordID: event.RecordID, Status: status, Action: event.Action, Actor: event.Actor, Detail: event.Detail, At: event.At})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].At < items[j].At })
	return items
}

func statusForAction(action string) SiteStatus {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "submit":
		return StatusInReview
	case "confirm":
		return StatusConfirmed
	case "archive":
		return StatusArchived
	default:
		return StatusDraft
	}
}

func LastAction(events []AuditEvent) string {
	if len(events) == 0 {
		return "created"
	}
	latest := events[0]
	for _, event := range events[1:] {
		if event.At >= latest.At {
			latest = event
		}
	}
	return latest.Action
}

func EventActors(events []AuditEvent) []string {
	actors := make([]string, 0, len(events))
	for _, event := range events {
		actors = append(actors, event.Actor)
	}
	return UniqueStrings(actors)
}

func EventActions(events []AuditEvent) []string {
	actions := make([]string, 0, len(events))
	for _, event := range events {
		actions = append(actions, event.Action)
	}
	return UniqueStrings(actions)
}

func EventCount(events []AuditEvent, action string) int {
	count := 0
	for _, event := range events {
		if strings.EqualFold(event.Action, action) {
			count++
		}
	}
	return count
}

func HasAction(events []AuditEvent, action string) bool {
	return EventCount(events, action) > 0
}

func FilterEvents(events []AuditEvent, action string) []AuditEvent {
	filtered := make([]AuditEvent, 0)
	for _, event := range events {
		if action == "" || strings.EqualFold(event.Action, action) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func EventSummary(events []AuditEvent) map[string]int {
	summary := map[string]int{}
	for _, event := range events {
		summary[event.Action]++
	}
	return summary
}
