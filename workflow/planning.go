package workflow

import (
	"fmt"
	"sitepreflight/model"
	"sort"
	"strings"
)

type PlanStep struct {
	ID          string
	Title       string
	Owner       string
	State       string
	Required    bool
	DependsOn   []string
	Description string
}

type MigrationPlan struct {
	ID        string
	BatchID   string
	Name      string
	Steps     []PlanStep
	CreatedBy string
	Version   int
	Published bool
}

func DefaultPlan(batchID string) MigrationPlan {
	return MigrationPlan{ID: "plan-" + batchID, BatchID: batchID, Name: "migration readiness " + batchID, Version: 1, Steps: []PlanStep{
		{ID: "inventory", Title: "inventory current site", State: "pending", Required: true, Description: "capture domains, owners and dependencies"},
		{ID: "dns", Title: "verify DNS", State: "pending", Required: true, DependsOn: []string{"inventory"}, Description: "confirm records and expected TTL"},
		{ID: "tls", Title: "verify TLS", State: "pending", Required: true, DependsOn: []string{"dns"}, Description: "confirm certificates and renewal path"},
		{ID: "redirects", Title: "check redirects", State: "pending", Required: true, DependsOn: []string{"tls"}, Description: "record redirect and canonical behavior"},
		{ID: "rollback", Title: "prepare rollback", State: "pending", Required: true, DependsOn: []string{"redirects"}, Description: "document a reversible fallback"},
		{ID: "handover", Title: "handover evidence", State: "pending", Required: false, DependsOn: []string{"rollback"}, Description: "share the final evidence packet"},
	}}
}

func (p MigrationPlan) Validate() error {
	if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.BatchID) == "" {
		return fmt.Errorf("plan identity required")
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("plan has no steps")
	}
	seen := map[string]bool{}
	for _, step := range p.Steps {
		if step.ID == "" || step.Title == "" {
			return fmt.Errorf("plan step identity required")
		}
		if seen[step.ID] {
			return fmt.Errorf("duplicate plan step %s", step.ID)
		}
		seen[step.ID] = true
	}
	for _, step := range p.Steps {
		for _, dependency := range step.DependsOn {
			if !seen[dependency] {
				return fmt.Errorf("step %s depends on unknown step %s", step.ID, dependency)
			}
		}
	}
	return nil
}

func (p MigrationPlan) Step(id string) (PlanStep, bool) {
	for _, step := range p.Steps {
		if step.ID == id {
			return step, true
		}
	}
	return PlanStep{}, false
}

func (p MigrationPlan) ReadyStep(id string) bool {
	step, ok := p.Step(id)
	if !ok || step.State == "done" || step.State == "blocked" {
		return false
	}
	for _, dependency := range step.DependsOn {
		prior, found := p.Step(dependency)
		if !found || prior.State != "done" {
			return false
		}
	}
	return true
}

func (p MigrationPlan) Complete() bool {
	for _, step := range p.Steps {
		if step.Required && step.State != "done" {
			return false
		}
	}
	return true
}

func (p MigrationPlan) Progress() (int, int) {
	done := 0
	required := 0
	for _, step := range p.Steps {
		if step.Required {
			required++
			if step.State == "done" {
				done++
			}
		}
	}
	return done, required
}

func (p MigrationPlan) CompletionPercent() int {
	done, required := p.Progress()
	if required == 0 {
		return 0
	}
	return done * 100 / required
}

func (p MigrationPlan) Pending() []PlanStep {
	pending := make([]PlanStep, 0)
	for _, step := range p.Steps {
		if step.State != "done" {
			pending = append(pending, step)
		}
	}
	return pending
}

func (p MigrationPlan) Available() []PlanStep {
	available := make([]PlanStep, 0)
	for _, step := range p.Steps {
		if p.ReadyStep(step.ID) {
			available = append(available, step)
		}
	}
	return available
}

func (p MigrationPlan) SetState(id, state string) (MigrationPlan, error) {
	if state != "pending" && state != "done" && state != "blocked" {
		return p, fmt.Errorf("invalid plan state %s", state)
	}
	if _, ok := p.Step(id); !ok {
		return p, fmt.Errorf("unknown plan step %s", id)
	}
	copy := p
	copy.Steps = append([]PlanStep(nil), p.Steps...)
	for i := range copy.Steps {
		if copy.Steps[i].ID == id {
			copy.Steps[i].State = state
		}
	}
	return copy, nil
}

func (p MigrationPlan) Assign(id, owner string) (MigrationPlan, error) {
	if strings.TrimSpace(owner) == "" {
		return p, fmt.Errorf("plan owner required")
	}
	copy := p
	copy.Steps = append([]PlanStep(nil), p.Steps...)
	for i := range copy.Steps {
		if copy.Steps[i].ID == id {
			copy.Steps[i].Owner = owner
			return copy, nil
		}
	}
	return p, fmt.Errorf("unknown plan step %s", id)
}

func (p MigrationPlan) RequiredSteps() []PlanStep {
	steps := make([]PlanStep, 0)
	for _, step := range p.Steps {
		if step.Required {
			steps = append(steps, step)
		}
	}
	return steps
}

func (p MigrationPlan) Owners() []string {
	owners := make([]string, 0)
	for _, step := range p.Steps {
		if step.Owner != "" {
			owners = append(owners, step.Owner)
		}
	}
	return model.UniqueStrings(owners)
}

func (p MigrationPlan) SortSteps() MigrationPlan {
	copy := p
	copy.Steps = append([]PlanStep(nil), p.Steps...)
	sort.SliceStable(copy.Steps, func(i, j int) bool { return copy.Steps[i].ID < copy.Steps[j].ID })
	return copy
}

func (p MigrationPlan) Summary() string {
	done, total := p.Progress()
	return fmt.Sprintf("%s %d/%d required steps complete", p.BatchID, done, total)
}

func PlanForRecords(records []model.Record) []MigrationPlan {
	plans := make([]MigrationPlan, 0)
	for _, batch := range model.BatchIDs(records) {
		plans = append(plans, DefaultPlan(batch))
	}
	return plans
}

func ValidatePlans(plans []MigrationPlan) error {
	seen := map[string]bool{}
	for _, plan := range plans {
		if err := plan.Validate(); err != nil {
			return err
		}
		if seen[plan.BatchID] {
			return fmt.Errorf("duplicate plan batch %s", plan.BatchID)
		}
		seen[plan.BatchID] = true
	}
	return nil
}

func NextRequiredStep(p MigrationPlan) (PlanStep, bool) {
	available := p.Available()
	for _, step := range available {
		if step.Required {
			return step, true
		}
	}
	return PlanStep{}, false
}

func BlockedSteps(p MigrationPlan) []PlanStep {
	blocked := make([]PlanStep, 0)
	for _, step := range p.Steps {
		if step.State == "blocked" {
			blocked = append(blocked, step)
		}
	}
	return blocked
}

func PlanStateCounts(p MigrationPlan) map[string]int {
	counts := map[string]int{"pending": 0, "done": 0, "blocked": 0}
	for _, step := range p.Steps {
		counts[step.State]++
	}
	return counts
}
