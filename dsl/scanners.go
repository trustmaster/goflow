package dsl

import (
	"bytes"
	"regexp"
	"strings"
)

// ScanChars scans a token withing a given set of characters
type ScanChars struct {
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

// Process reads IIPs and validates incoming tokens
func (s *ScanChars) Process() {
	// Read IIPs first
	set := ""
	tokenType := ""
	ok := true
	if set, ok = <-s.Set; !ok {
		return
	}
	var matcher func(r rune) bool
	var reg *regexp.Regexp
	if set[0] == '[' && set[len(set)-1] == ']' {
		// A regexp class
		var err error
		reg, err = regexp.Compile(set)
		if err != nil {
			// TODO error handling
			return
		}
		matcher = func(r rune) bool {
			return reg.Match([]byte{byte(r)})
		}
	} else {
		// Replace special chars
		set = strings.ReplaceAll(set, `\t`, "\t")
		set = strings.ReplaceAll(set, `\r`, "\r")
		set = strings.ReplaceAll(set, `\n`, "\n")
		matcher = func(r rune) bool {
			return strings.ContainsRune(set, r)
		}
	}

	if tokenType, ok = <-s.Type; !ok {
		return
	}
	// Then process the incoming tokens
	for tok := range s.In {
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
		if buf.Len() > 0 {
			tok.Type = TokenType(tokenType)
			tok.Value = buf.String()
			s.Hit <- tok
		} else {
			s.Miss <- tok
		}
	}
}

// ScanKeyword scans a specific keyword that is not part of other word
type ScanKeyword struct {
	// Word is a case-insensitive keyword
	Word <-chan string
	// Type is an IIP that contains string token type associated with the set
	Type <-chan string
	// In is an incoming empty token
	In <-chan Token
	// Hit is a successfully matched token
	Hit chan<- Token
	// Miss is the unmodified empty token in case there was no match
	Miss chan<- Token
}

// Process reads IIPs and validates incoming tokens
func (s *ScanKeyword) Process() {
	// Read IIPs
	word := ""
	tokenType := ""
	ok := true
	if word, ok = <-s.Word; !ok {
		return
	}
	word = strings.ToUpper(word)
	wordLen := len(word)
	if tokenType, ok = <-s.Type; !ok {
		return
	}
	keywordReg := regexp.MustCompile(`[\w_]`)
	// Process incoming tokens
	for tok := range s.In {
		dataLen := len(tok.File.Data)
		if tok.Pos+wordLen > dataLen {
			// Data is too short
			s.Miss <- tok
			continue
		}
		if strings.ToUpper(string(tok.File.Data[tok.Pos:tok.Pos+wordLen])) == word {
			// Potential match, should be followed by EOF or non-word character
			if tok.Pos+wordLen < dataLen {
				nextChar := tok.File.Data[tok.Pos+wordLen]
				if keywordReg.Match([]byte{nextChar}) {
					// This is not the whole word
					s.Miss <- tok
					continue
				}
			}
			// Checks passed, it's a match
			tok.Type = TokenType(tokenType)
			tok.Value = word
			s.Hit <- tok
		}
		// No match
		s.Miss <- tok
	}
}

// ScanSequence scans a sequence of characters
type ScanSequence struct {
	// Seq is the exact sequence of characters to read
	Seq <-chan string
	// Type is an IIP that contains string token type associated with the set
	Type <-chan string
	// In is an incoming empty token
	In <-chan Token
	// Hit is a successfully matched token
	Hit chan<- Token
	// Miss is the unmodified empty token in case there was no match
	Miss chan<- Token
}

// ScanQuoted scans a quoted string
type ScanQuoted struct {
	// In is an incoming empty token
	In <-chan Token
	// Hit is a successfully matched token
	Hit chan<- Token
	// Miss is the unmodified empty token in case there was no match
	Miss chan<- Token
}

// ScanInvalid returns an illegal token
type ScanInvalid struct {
	// In is an incoming empty token
	In <-chan Token
	// Token returns and invalid token
	Token chan<- Token
}
