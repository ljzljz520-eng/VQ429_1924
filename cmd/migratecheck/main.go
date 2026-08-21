package main

import (
	"flag"
	"fmt"
	"os"
	"sitepreflight/model"
	"sitepreflight/registry"
	"sitepreflight/report"
	"sitepreflight/search"
	"sitepreflight/storage"
)

func main() {
	path := flag.String("db", "preflight.db", "database path")
	cmd := flag.String("command", "list", "list,add,export")
	id := flag.String("id", "demo-1", "record id")
	batch := flag.String("batch", "429-21", "batch id")
	flag.Parse()
	s, e := storage.Open(*path)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	defer s.Close()
	switch *cmd {
	case "add":
		e = registry.New(s).Register(model.Record{ID: *id, BatchID: *batch, Name: "demo site", Owner: "ops", Domain: "example.test", Status: model.StatusDraft, Checklist: []string{"dns", "tls"}, Tags: []string{"pilot"}})
	case "list":
		var rs []model.Record
		rs, e = search.New(s).Find(model.Filter{})
		for _, r := range rs {
			fmt.Printf("%s %s %s\n", r.ID, r.Status, r.Domain)
		}
	case "export":
		e = report.New(s).CSV(os.Stdout, model.Filter{BatchID: *batch, ConfirmedOnly: true})
	default:
		e = fmt.Errorf("unknown command")
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
