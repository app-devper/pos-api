// Package legacy migrates the database of the 2024 build (no Branches) so the
// current API can see its documents.
package legacy

import "go.mongodb.org/mongo-driver/bson/primitive"

// Change is a field the migration sets (apply) or would set (report) on Count
// documents of one collection.
type Change struct {
	Collection string
	Field      string
	Count      int64
}

// Finding is a problem the migration reports but does not fix.
type Finding struct {
	Kind   string
	Detail string
	Ids    []string // at most 20, hex
	Total  int64
}

// Report is what a Migrate run did, or in report mode would do.
type Report struct {
	HQ        primitive.ObjectID
	HQCreated bool
	Changes   []Change
	Findings  []Finding
}

// HasFindings reports whether anything needs a person's attention.
func (r Report) HasFindings() bool {
	return len(r.Findings) > 0
}

// withChanges is a copy of r with changes in place of its Changes.
func (r Report) withChanges(changes []Change) Report {
	r.Changes = changes
	return r
}
