package dsl

// Statement is a trivia-free sequence of tokens that forms one DSL statement.
type Statement struct {
	Tokens []Token
	Span   Span
}

func newStatement(tokens []Token) Statement {
	statementTokens := append([]Token(nil), tokens...)
	span := statementTokens[0].Span
	span.End = statementTokens[len(statementTokens)-1].Span.End

	return Statement{
		Tokens: statementTokens,
		Span:   span,
	}
}
