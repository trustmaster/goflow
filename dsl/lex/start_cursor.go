package lex

import "github.com/trustmaster/goflow/dsl/types"

// StartCursor starts a cursor stream from an input file.
type StartCursor struct {
	File <-chan *types.File
	Out  chan<- types.Cursor
}

// Process emits an initial cursor for each incoming file.
func (s *StartCursor) Process() {
	for file := range s.File {
		if file == nil {
			continue
		}

		s.Out <- types.NewCursor(file)
	}
}
