package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestScanners(t *testing.T) {
	type testCase struct {
		c       string
		name    string
		set     string
		tokType TokenType
		data    string
		hit     bool
		value   string
		pos     int
	}

	cases := []testCase{
		// ScanChars
		{
			c:       "dsl/ScanChars",
			name:    "Matching an integer",
			set:     "0123456789",
			tokType: tokInt,
			data:    "123.456",
			pos:     0,
			hit:     true,
			value:   "123",
		},
		{
			c:       "dsl/ScanChars",
			name:    "Not matching an integer",
			set:     "0123456789",
			tokType: tokInt,
			data:    "a123456",
			pos:     0,
			hit:     false,
			value:   "",
		},
		{
			c:       "dsl/ScanChars",
			name:    "Matching a regexp class",
			set:     "[a-zA-Z0-9_]",
			tokType: tokIdent,
			data:    "cool_Var123.PORT ->",
			pos:     0,
			hit:     true,
			value:   "cool_Var123",
		},
		{
			c:       "dsl/ScanChars",
			name:    "Not matching a regexp class",
			set:     `[ \t\r\n]`,
			tokType: tokWhitespace,
			data:    "Illeg al\n",
			pos:     0,
			hit:     false,
			value:   "",
		},
		{
			c:       "dsl/ScanChars",
			name:    "Matching with non-zero offset",
			set:     ".",
			tokType: tokDot,
			data:    "abc.def",
			pos:     3,
			hit:     true,
			value:   ".",
		},
		// ScanKeyword
		{
			c:       "dsl/ScanKeyword",
			name:    "Matching a keyword",
			set:     "INPORT",
			tokType: tokInport,
			data:    "INPORT=Foo.BAR:BAR",
			pos:     0,
			hit:     true,
			value:   "INPORT",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Matching a keyword (case-insensitive)",
			set:     "inport",
			tokType: tokInport,
			data:    "inPort=Foo.BAR:BAR",
			pos:     0,
			hit:     true,
			value:   "INPORT",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Matching a keyword with non-zero offset",
			set:     "INPORT",
			tokType: tokInport,
			data:    ":FOO\nINPORT=Foo.BAR:BAR",
			pos:     5,
			hit:     true,
			value:   "INPORT",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Input too short",
			set:     "INPORT",
			tokType: tokInport,
			data:    "INP",
			pos:     0,
			hit:     false,
			value:   "",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Is not a whole word",
			set:     "INPORT",
			tokType: tokInport,
			data:    "INPORTSTONE=Foo.BAR:BAR",
			pos:     0,
			hit:     false,
			value:   "",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Does not match",
			set:     "INPORT",
			tokType: tokInport,
			data:    "OUTPORT=Foo.BAR:BAR",
			pos:     0,
			hit:     false,
			value:   "",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Matches a single char operator",
			set:     ".",
			tokType: tokDot,
			data:    "Foo.BAR",
			pos:     3,
			hit:     true,
			value:   ".",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Matches a multi char operator",
			set:     "->",
			tokType: tokArrow,
			data:    "FOO -> BAR",
			pos:     4,
			hit:     true,
			value:   "->",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Does not match absent operator",
			set:     ":",
			tokType: tokColon,
			data:    "Foo.BAR",
			pos:     0,
			hit:     false,
			value:   "",
		},
		{
			c:       "dsl/ScanKeyword",
			name:    "Does not match part of a longer oprator",
			set:     "->",
			tokType: tokArrow,
			data:    "FOO ->? BAR",
			pos:     4,
			hit:     false,
			value:   "",
		},
		// ScanComment
		{
			c:       "dsl/ScanComment",
			name:    "Scans a comment till the end of data",
			set:     "#",
			tokType: tokComment,
			data:    "Foo BAR -> BOO Baz # This is a comment",
			pos:     19,
			hit:     true,
			value:   "# This is a comment",
		},
		{
			c:       "dsl/ScanComment",
			name:    "Scans a comment till the end of line",
			set:     "#",
			tokType: tokComment,
			data:    "Foo BAR -> BOO Baz # This is a comment\r\nNew LINE",
			pos:     19,
			hit:     true,
			value:   "# This is a comment",
		},
		{
			c:       "dsl/ScanComment",
			name:    "Does not match non-comment",
			set:     "#",
			tokType: tokComment,
			data:    "Foo BAR -> BOO Baz",
			pos:     0,
			hit:     false,
			value:   "",
		},
		// ScanQuoted
		{
			c:       "dsl/ScanQuoted",
			name:    "Scans a quoted string",
			set:     "'",
			tokType: tokQuotedStr,
			data:    `'This is an IIP' -> IN Foo`,
			pos:     0,
			hit:     true,
			value:   "'This is an IIP'",
		},
		{
			c:       "dsl/ScanQuoted",
			name:    "Supports escaping of quote char",
			set:     `"`,
			tokType: tokQuotedStr,
			data:    `"This is an \"IIP\"" -> IN Foo`,
			pos:     0,
			hit:     true,
			value:   `"This is an "IIP""`,
		},
		{
			c:       "dsl/ScanQuoted",
			name:    "Supports escaping of backslash itself",
			set:     `"`,
			tokType: tokQuotedStr,
			data:    `"This is an IIP\\" -> IN Foo`,
			pos:     0,
			hit:     true,
			value:   `"This is an IIP\"`,
		},
		{
			c:       "dsl/ScanQuoted",
			name:    "Does not work without quote char",
			set:     "",
			tokType: tokQuotedStr,
			data:    `"This is an IIP" -> IN Foo`,
			pos:     0,
			hit:     false,
			value:   "",
		},
		{
			c:       "dsl/ScanQuoted",
			name:    "Does not match if no quote found",
			set:     "'",
			tokType: tokQuotedStr,
			data:    `This is not an IIP -> IN Foo`,
			pos:     0,
			hit:     false,
			value:   "",
		},
		{
			c:       "dsl/ScanQuoted",
			name:    "Captures everything if quote is not closed",
			set:     `"`,
			tokType: tokQuotedStr,
			data:    `I "Forgot to close -> IN Foo`,
			pos:     2,
			hit:     true,
			value:   `"Forgot to close -> IN Foo`,
		},
	}

	runCase := func(tc testCase, t *testing.T) {
		set := make(chan string, 1)     // for non-blocking send
		tokType := make(chan string, 1) // for non-blocking send
		in := make(chan Token)
		out := make(chan Token)

		f := goflow.NewFactory()
		if err := RegisterComponents(f); err != nil {
			t.Error(err)
			return
		}

		i, err := f.Create(tc.c)
		if err != nil {
			t.Error(err)
			return
		}
		s := i.(scanner)
		s.assign(Scanner{
			Set:  set,
			Type: tokType,
			In:   in,
			Out:  out,
		})

		wait := goflow.Run(s)

		tokType <- string(tc.tokType)
		set <- tc.set
		in <- Token{
			File: &File{
				Name: "test.fbp",
				Data: []byte(tc.data),
			},
			Pos: tc.pos,
		}

		go func() {
			tok, ok := <-out
			if !ok {
				return
			}
			if tok.Type == tokIllegal {
				if tc.hit {
					t.Errorf("Unexpected miss: '%s' at %d", tok.Value, tok.Pos)
				}
			} else {
				if !tc.hit {
					t.Errorf("Unexpected hit: '%s' at %d", tok.Value, tok.Pos)
				}
				if tok.Type != tc.tokType {
					t.Errorf("Unexpected token type: %s", tok.Type)
				}
				if tok.Value != tc.value {
					t.Errorf("Unexpected token value: '%s'", tok.Value)
				}
			}
			close(in)
		}()

		<-wait
	}

	t.Parallel()
	for _, c := range cases {
		t.Run(c.c+": "+c.name, func(t *testing.T) {
			runCase(c, t)
		})
	}
}
