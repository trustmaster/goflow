package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestCollect(t *testing.T) {
	type input struct {
		tokType TokenType
		value   string
	}

	type testCase struct {
		name          string
		data          string
		pos           int
		inputs        []input
		expectedType  TokenType
		expectedValue string
		hasNext       bool
		nextPos       int
	}

	cases := []testCase{
		{
			name: "One of three inputs matched correctly",
			data: "IN Collect(dsl/Collect) OUT",
			pos:  0,
			inputs: []input{
				{
					tokType: tokIllegal,
					value:   "",
				},
				{
					tokType: tokIdent,
					value:   "IN",
				},
				{
					tokType: tokIllegal,
					value:   "",
				},
			},
			expectedType:  tokIdent,
			expectedValue: "IN",
			hasNext:       true,
			nextPos:       2,
		},
		{
			name: "None of three inputs matched",
			data: "IN Collect(dsl/Collect) OUT",
			pos:  3,
			inputs: []input{
				{
					tokType: tokIllegal,
					value:   "",
				},
				{
					tokType: tokIllegal,
					value:   "",
				},
				{
					tokType: tokIllegal,
					value:   "",
				},
			},
			expectedType:  tokIllegal,
			expectedValue: "",
			hasNext:       false,
			nextPos:       3,
		},
		{
			name: "Two of three inputs matched correctly, first match returned",
			data: "IN Inport(dsl/Collect) OUT",
			pos:  3,
			inputs: []input{
				{
					tokType: tokInport,
					value:   "Inport",
				},
				{
					tokType: tokIdent,
					value:   "Inport",
				},
				{
					tokType: tokIllegal,
					value:   "",
				},
			},
			expectedType:  tokInport,
			expectedValue: "Inport",
			hasNext:       true,
			nextPos:       9,
		},
	}

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Error(err)
		return
	}

	t.Parallel()

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			i, err := f.Create("dsl/Collect")
			if err != nil {
				t.Error(err)
				return
			}
			collect := i.(*Collect)

			ins := make([](chan Token), len(c.inputs))
			out := make(chan Token)
			next := make(chan Token)
			collect.Out = out
			collect.Next = next
			collect.In = make([](<-chan Token), len(c.inputs))

			for i := range c.inputs {
				ins[i] = make(chan Token, 1)
				collect.In[i] = ins[i]
			}

			wait := goflow.Run(collect)

			for i, in := range c.inputs {
				i, in := i, in
				go func() {
					ins[i] <- Token{
						File: &File{
							Name: "test.fbp",
							Data: []byte(c.data),
						},
						Pos:   c.pos,
						Type:  in.tokType,
						Value: in.value,
					}
				}()
			}

			go func() {
				res := <-out
				if res.Type != c.expectedType {
					t.Errorf("Expected type '%s', got '%s'", c.expectedType, res.Type)
				}
				if res.Value != c.expectedValue {
					t.Errorf("Expected value '%s', got '%s'", c.expectedValue, res.Value)
				}

				if c.hasNext {
					nextTok := <-next

					if nextTok.Pos != c.nextPos {
						t.Errorf("Expected next pos %d, got %d", c.nextPos, nextTok.Pos)
					}
				}

				for i := range c.inputs {
					close(ins[i])
				}
			}()

			<-wait
		})
	}
}
