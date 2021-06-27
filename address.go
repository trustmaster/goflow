package goflow

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// address is a full port accessor including the index part.
type address struct {
	proc  string // Process name
	port  string // Component port name
	key   string // Port key (only for map ports)
	index int    // Port index (only for array ports)
}

// noIndex is a "zero" index value. Not a `0` since 0 is a valid array index.
const noIndex = -1

type portKind uint

const (
	portKindNone portKind = iota
	portKindChan
	portKindArray
	portKindMap
)

func (a address) kind() portKind {
	switch {
	case len(a.proc) == 0 || len(a.port) == 0:
		return portKindNone
	case a.index != noIndex:
		return portKindArray
	case len(a.key) != 0:
		return portKindMap
	default:
		return portKindChan
	}
}

func (a address) String() string {
	switch a.kind() {
	case portKindChan:
		return fmt.Sprintf("%s.%s", a.proc, a.port)
	case portKindArray:
		return fmt.Sprintf("%s.%s[%d]", a.proc, a.port, a.index)
	case portKindMap:
		return fmt.Sprintf("%s.%s[%s]", a.proc, a.port, a.key)
	case portKindNone: // makes go-lint happy
	}

	return "<none>"
}

// parseAddress validates and constructs a port address.
// port parameter may include an array index ("<port name>[<index>]") or a hashmap key ("<port name>[<key>]").
func parseAddress(proc, port string) (address, error) {
	switch {
	case len(proc) == 0:
		return address{}, fmt.Errorf("empty process name")
	case len(port) == 0:
		return address{}, fmt.Errorf("empty port name")
	}

	// Validate the proc contents
	for i, r := range proc {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return address{}, fmt.Errorf("unexpected %q at process name index %d", r, i)
		}
	}

	keyPos := 0
	a := address{
		proc:  proc,
		port:  port,
		index: noIndex,
	}

	// Validate and parse the port contents in one scan
	for i, r := range port {
		switch {
		case r == '[':
			if i == 0 || keyPos > 0 {
				// '[' at the very beginning of the port or a second '[' found
				return address{}, fmt.Errorf("unexpected '[' at port name index %d", i)
			}

			keyPos = i + 1
			a.port = port[:i]
		case r == ']':
			switch {
			case keyPos == 0:
				// No preceding matching '['
				return address{}, fmt.Errorf("unexpected ']' at port name index %d", i)
			case i != len(port)-1:
				// Closing bracket is not the last rune
				return address{}, fmt.Errorf("unexpected %q at port name index %d", port[i+1:], i)
			}

			if idx, err := strconv.Atoi(port[keyPos:i]); err != nil {
				a.key = port[keyPos:i]
			} else {
				a.index = idx
			}
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			return address{}, fmt.Errorf("unexpected %q at port name index %d", r, i)
		}
	}

	if keyPos != 0 && len(a.key) == 0 && a.index == noIndex {
		return address{}, fmt.Errorf("unmatched '[' at port name index %d", keyPos-1)
	}

	a.port = capitalizePortName(a.port)

	return a, nil
}

// capitalizePortName converts port names defined in UPPER or lower case to Title case,
// which is more common for structs in Go.
func capitalizePortName(name string) string {
	lower := strings.ToLower(name)
	upper := strings.ToUpper(name)

	if name == lower || name == upper {
		return strings.Title(lower)
	}

	return name
}
