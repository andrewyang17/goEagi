// Package goEagi of eagi.go provides an Eagi type,
// which composited *agi.Session (an external package) and
// some member fields to be used across the program.

package goEagi

import (
	"github.com/pkg/errors"
	"github.com/zaf/agi"
)

// Eagi composited type *agi.Session
// and provides extra information (based on configuration)
// to the context.
type Eagi struct {
	*agi.Session

	Id        string // Unique id
	Extension string // Extension number
	Actor     string // Customer or Agent
}

func New() (*Eagi, error) {
	newAgi := agi.New()
	if err := newAgi.Init(nil); err != nil {
		return nil, errors.Wrap(err, "failed to initialize agi session")
	}

	e := Eagi{}
	e.Session = newAgi

	env := e.Session.Env
	e.Extension = env["arg_1"]
	e.Id = env["arg_2"]

	return &e, nil
}
