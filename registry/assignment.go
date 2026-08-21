package registry

import (
	"fmt"
	"sitepreflight/model"
)

func (c *Catalog) AssignOwner(id, owner string) error {
	record, err := c.store.GetRecord(id)
	if err != nil {
		return err
	}
	if owner == "" {
		return fmt.Errorf("owner required")
	}
	record.Owner = owner
	if err := c.store.PutRecord(record); err != nil {
		return err
	}
	return c.store.PutAudit(model.AuditEvent{ID: id + "-owner", RecordID: id, Actor: owner, Action: "assign", Detail: "owner assigned"})
}

func (c *Catalog) AddChecklist(id, item string) error {
	record, err := c.store.GetRecord(id)
	if err != nil {
		return err
	}
	record = model.AddChecklistItem(record, item)
	return c.store.PutRecord(record)
}

func (c *Catalog) RemoveChecklist(id, item string) error {
	record, err := c.store.GetRecord(id)
	if err != nil {
		return err
	}
	record = model.RemoveChecklistItem(record, item)
	return c.store.PutRecord(record)
}

func (c *Catalog) AddTag(id, tag string) error {
	record, err := c.store.GetRecord(id)
	if err != nil {
		return err
	}
	return c.store.PutRecord(model.AddTag(record, tag))
}

func (c *Catalog) RemoveTag(id, tag string) error {
	record, err := c.store.GetRecord(id)
	if err != nil {
		return err
	}
	return c.store.PutRecord(model.RemoveTag(record, tag))
}

func (c *Catalog) Normalize(id string) (model.Record, error) {
	record, err := c.store.GetRecord(id)
	if err != nil {
		return model.Record{}, err
	}
	record = model.NormalizeRecord(record)
	return record, c.store.PutRecord(record)
}
