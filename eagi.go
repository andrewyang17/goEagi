// Package goEagi of eagi.go provides an Eagi type,
// which composited *agi.Session (an external package) and
// some member fields to be used across the program.

package goEagi

import (
	"fmt"
	"github.com/zaf/agi"
)

// Eagi composited type *agi.Session
// and provides extra information (based on configuration)
// to the context.
type Eagi struct {
	*agi.Session
}

func New() (*Eagi, error) {
	newAgi := agi.New()
	if err := newAgi.Init(nil); err != nil {
		return nil, fmt.Errorf("failed to initialize eagi session: %v\n", err)
	}

	e := Eagi{}
	e.Session = newAgi

	return &e, nil
}
