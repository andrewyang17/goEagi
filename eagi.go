// Package goEagi of eagi.go provides an Eagi type,
// which composited *agi.Session (an external package) and
// some member fields to be used across the program (upcoming milestones).

package goEagi

import (
	"fmt"

	"github.com/zaf/agi"
)

type Eagi struct {
	*agi.Session
}

func New() (*Eagi, error) {
	newSession := agi.New()
	if err := newSession.Init(nil); err != nil {
		return nil, fmt.Errorf("failed to initialize eagi session: %v\n", err)
	}

	e := Eagi{}
	e.Session = newSession

	return &e, nil
}
