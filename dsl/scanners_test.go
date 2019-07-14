package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestScanSet(t *testing.T) {
	type testCase struct {
		name    string
		set     string
		tokType TokenType
		data    string
		hit     bool
		value   string
	}

	cases := []testCase{
		{
			name:    "Matching an integer",
			set:     "0123456789",
			tokType: tokInt,
			data:    "123.456",
			hit:     true,
			value:   "123",
		},
		{
			name:    "Not matching an integer",
			set:     "0123456789",
			tokType: tokInt,
			data:    "a123456",
			hit:     false,
			value:   "",
		},
		{
			name:    "Matching a regexp class",
			set:     "[a-zA-Z0-9_]",
			tokType: tokIdent,
			data:    "cool_Var123.PORT ->",
			hit:     true,
			value:   "cool_Var123",
		},
		{
			name:    "Not matching a regexp class",
			set:     `[ \t\r\n]`,
			tokType: tokWhitespace,
			data:    "Illeg al\n",
			hit:     false,
			value:   "",
		},
	}

	runCase := func(tc testCase, t *testing.T) {
		set := make(chan string, 1)     // for non-blocking send
		tokType := make(chan string, 1) // for non-blocking send
		in := make(chan Token)
		hit := make(chan Token)
		miss := make(chan Token)

		f := goflow.NewFactory()
		if err := RegisterComponents(f); err != nil {
			t.Error(err)
			return
		}

		i, err := f.Create("dsl/ScanChars")
		if err != nil {
			t.Error(err)
			return
		}
		c := i.(*ScanChars)
		c.Set = set
		c.Type = tokType
		c.In = in
		c.Hit = hit
		c.Miss = miss

		wait := goflow.Run(c)

		tokType <- string(tc.tokType)
		set <- tc.set
		in <- Token{
			File: &File{
				Name: "test.fbp",
				Data: []byte(tc.data),
			},
			Pos: 0,
		}

		go func() {
			select {
			case tok, ok := <-hit:
				if !ok {
					return
				}
				if !tc.hit {
					t.Errorf("Unexpected hit: %v", tok)
				}
				if tok.Type != tc.tokType {
					t.Errorf("Unexpected token type: %s", tok.Type)
				}
				if tok.Value != tc.value {
					t.Errorf("Unexpected token value: %s", tok.Value)
				}
			case tok, ok := <-miss:
				if !ok {
					return
				}
				if tc.hit {
					t.Errorf("Unexpected miss: %v", tok)
				}
			}

			close(in)
		}()

		<-wait
	}

	t.Parallel()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			runCase(c, t)
		})
	}
}
