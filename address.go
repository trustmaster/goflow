package goflow

import (
	"fmt"
	"strconv"
	"strings"
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

func (a address) String() string {
	if a.key != "" {
		return fmt.Sprintf("%s.%s[%s]", a.proc, a.port, a.key)
	}

	return fmt.Sprintf("%s.%s", a.proc, a.port)
}

// parseAddress unfolds a string port name into parts, including array index or hashmap key.
func parseAddress(proc, port string) address {
	n := address{
		proc:  proc,
		port:  port,
		index: noIndex,
	}
	keyPos := 0
	key := ""

	for i, r := range port {
		if r == '[' {
			keyPos = i + 1
			n.port = port[0:i]
		}

		if r == ']' {
			key = port[keyPos:i]
		}
	}

	n.port = capitalizePortName(n.port)

	if key == "" {
		return n
	}

	if i, err := strconv.Atoi(key); err == nil {
		n.index = i
	} else {
		n.key = key
	}

	n.key = key

	return n
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

type addressKind uint

const (
	addressKindNone addressKind = iota
	addressKindChan
	addressKindArray
	addressKindMap
)

func (a address) kind() addressKind {
	switch {
	case len(a.proc) == 0 || len(a.port) == 0:
		return addressKindNone
	case a.index != noIndex:
		return addressKindArray
	case len(a.key) != 0:
		return addressKindMap
	default:
		return addressKindChan
	}
}
