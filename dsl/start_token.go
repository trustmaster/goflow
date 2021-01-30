package dsl

// StartToken starts a stream of tokens from a given file.
type StartToken struct {
	File <-chan *File
	Init chan<- Token // Initial token, doesn't need scanning
	Next chan<- Token // Next token to scan
}

// Process reads files and starts token streams.
func (s *StartToken) Process() {
	for f := range s.File {
		t := Token{
			Type:  tokNewFile,
			File:  f,
			Pos:   0,
			Value: f.Name,
		}
		// Send to Init, which should go directly to Tokenizer output
		s.Init <- t
		// Send to Next, which should go to Scanners
		s.Next <- t
	}
}
