package dsl

import (
	"bytes"
	"regexp"
	"strings"
)

// Scanner is a unified structure for scanner components which defines their port signature
type Scanner struct {
	// Set is an IIP that contains valid characters. Supports special characters: \r, \n, \t.
	// A regular expression character class can be passed like: "[a-zA-Z\s]".
	Set <-chan string
	// Type is an IIP that contains string token type associated with the set
	Type <-chan string
	// In is an incoming empty token
	In <-chan Token
	// Hit is a successfully matched token
	Hit chan<- Token
	// Miss is the unmodified empty token in case there was no match
	Miss chan<- Token
}

// scanner is used to test Scanner components via common interface
type scanner interface {
	assign(Scanner)
	Process()
}

// assign binds ports for testing
func (s *Scanner) assign(ports Scanner) {
	s.Set = ports.Set
	s.Type = ports.Type
	s.In = ports.In
	s.Hit = ports.Hit
	s.Miss = ports.Miss
}

// readIIPs reads configuration ports. Connections have to be buffered to avoid order deadlock
func (s *Scanner) readIIPs() (set string, tokenType string, ok bool) {
	if set, ok = <-s.Set; !ok {
		return
	}
	if tokenType, ok = <-s.Type; !ok {
		return
	}
	return
}

// scanTok is a callback that scans a single token
type scanTok func(Token) (Token, bool)

// handleTokens reads incoming tokens and applies a scan callback to them
func (s *Scanner) handleTokens(scan scanTok) {
	// Read incoming tokens and scan them with a callback
	for tok := range s.In {
		t, match := scan(tok)
		if match {
			s.Hit <- t
		} else {
			s.Miss <- t
		}
	}
}

// ScanChars scans a token of characters belonging to Set
type ScanChars struct {
	Scanner
}

// Process reads IIPs and validates incoming tokens
func (s *ScanChars) Process() {
	set, tokenType, ok := s.readIIPs()
	if !ok {
		return
	}
	matcher := s.matcher(set)
	s.handleTokens(func(tok Token) (Token, bool) {
		buf := bytes.NewBufferString("")
		dataLen := len(tok.File.Data)
		// Read as many chars within the set as possible
		for i := tok.Pos; i < dataLen; i++ {
			r := rune(tok.File.Data[i])
			if r == rune(0) || !matcher(r) {
				break
			}
			buf.WriteRune(r)
		}
		tok.Value = buf.String()
		tok.Type = TokenType(tokenType)
		return tok, buf.Len() > 0
	})
}

func (s *ScanChars) matcher(set string) func(r rune) bool {
	if set[0] == '[' && set[len(set)-1] == ']' {
		// A regexp class
		var reg *regexp.Regexp
		var err error
		reg, err = regexp.Compile(set)
		if err != nil {
			// TODO error handling
			return func(r rune) bool {
				return false
			}
		}
		return func(r rune) bool {
			return reg.Match([]byte{byte(r)})
		}
	}
	// Replace special chars
	set = strings.ReplaceAll(set, `\t`, "\t")
	set = strings.ReplaceAll(set, `\r`, "\r")
	set = strings.ReplaceAll(set, `\n`, "\n")
	return func(r rune) bool {
		return strings.ContainsRune(set, r)
	}
}

// ScanKeyword scans a case-insensitive keyword that is not part of another word.
// If Set is an identifier, it makes sure the keyword is not a substring of another identifier.
// If Set is an operator, it makes sure the operator is followed by identifier or space.
type ScanKeyword struct {
	Scanner
}

// Process reads IIPs and validates incoming tokens
func (s *ScanKeyword) Process() {
	word, tokenType, ok := s.readIIPs()
	if !ok {
		return
	}
	word = strings.ToUpper(word)
	wordLen := len(word)

	identReg := regexp.MustCompile(`[\w_]`)
	shouldNotBeFollowedBy := identReg
	isIdent := identReg.MatchString(word)
	if !isIdent {
		shouldNotBeFollowedBy = regexp.MustCompile(`[^\w\s]`)
	}

	s.handleTokens(func(tok Token) (Token, bool) {
		dataLen := len(tok.File.Data)
		if tok.Pos+wordLen > dataLen {
			// Data is too short
			return tok, false
		}
		tok.Value = string(tok.File.Data[tok.Pos : tok.Pos+wordLen])
		if strings.ToUpper(tok.Value) == word {
			// Potential match, should be followed by EOF or non-word character
			if tok.Pos+wordLen < dataLen {
				nextChar := tok.File.Data[tok.Pos+wordLen]
				if shouldNotBeFollowedBy.Match([]byte{nextChar}) {
					// This is not the whole word
					return tok, false
				}
			}
			// Checks passed, it's a match
			tok.Type = TokenType(tokenType)
			tok.Value = word
			return tok, true
		}
		// No match
		return tok, false
	})
}

// ScanComment scans a comment from hash till the end of line
type ScanComment struct {
	Scanner
}

// Process reads IIPs and validates incoming tokens
func (s *ScanComment) Process() {
	prefix, tokenType, ok := s.readIIPs()
	if !ok {
		return
	}
	s.handleTokens(func(tok Token) (Token, bool) {
		if tok.File.Data[tok.Pos] != prefix[0] {
			return tok, false
		}
		buf := bytes.NewBufferString("")
		dataLen := len(tok.File.Data)
		// Read all characters till the end of the line
		for i := tok.Pos; i < dataLen; i++ {
			r := rune(tok.File.Data[i])
			if r == rune(0) || r == rune('\n') || r == rune('\r') {
				break
			}
			buf.WriteRune(r)
		}
		tok.Value = buf.String()
		tok.Type = TokenType(tokenType)
		return tok, true
	})
}

// ScanQuoted scans a quoted string
type ScanQuoted struct {
	Scanner
}

// Process scans for quoted strings in the incoming tokens
func (s *ScanQuoted) Process() {
	quotes, tokenType, ok := s.readIIPs()
	if !ok {
		return
	}
	s.handleTokens(func(tok Token) (Token, bool) {
		// Find the quote char
		var q rune = 0
		for _, b := range quotes {
			if rune(tok.File.Data[tok.Pos]) == b {
				q = b
				break
			}
		}
		if q == 0 {
			return tok, false
		}

		var e rune = '\\'
		escaped := 0
		buf := bytes.NewBufferString(string(q))
		dataLen := len(tok.File.Data)
		for i := tok.Pos + 1; i < dataLen; i++ {
			r := rune(tok.File.Data[i])
			if r == e {
				escaped = (escaped + 1) % 2
				if escaped == 1 {
					continue
				}
			}
			buf.WriteRune(r)
			if r == q && escaped == 0 {
				break
			}
			escaped = 0
		}
		tok.Value = buf.String()
		tok.Type = TokenType(tokenType)
		return tok, true
	})
}

// ScanInvalid returns an illegal token
type ScanInvalid struct {
	// In is an incoming empty token
	In <-chan Token
	// Token returns and invalid token
	Token chan<- Token
}

// Process marks all incoming tokens as invalid
func (s *ScanInvalid) Process() {
	for tok := range s.In {
		tok.Type = tokIllegal
		tok.Value = string(tok.File.Data[tok.Pos])
		s.Token <- tok
	}
}
