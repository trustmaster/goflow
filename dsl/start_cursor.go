package dsl

// StartCursor starts a cursor stream from an input file.
type StartCursor struct {
	File <-chan *File
	Out  chan<- Cursor
}

// Process emits an initial cursor for each incoming file.
func (s *StartCursor) Process() {
	for file := range s.File {
		if file == nil {
			continue
		}

		s.Out <- newCursor(file)
	}
}
