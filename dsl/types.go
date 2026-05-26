package dsl

// File represents a source file.
type File struct {
	Name string
	Data []byte
}

// Span identifies a byte range in a source file.
type Span struct {
	File   string
	Offset int
	Line   int
	Column int
	End    int
}

// Cursor points to the next unread position in a file.
type Cursor struct {
	File   *File
	Offset int
	Line   int
	Column int
}

// TokenType differentiates lexer tokens.
type TokenType string

const (
	// Generic token types.
	TokIllegal    = TokenType("illegal")
	TokNewFile    = TokenType("newFile") // legacy tokenizer support
	TokEOF        = TokenType("eof")
	TokWhitespace = TokenType("whitespace")
	TokEOL        = TokenType("eol")
	TokComment    = TokenType("comment")

	// Literals.
	TokIdent  = TokenType("ident")
	TokInt    = TokenType("int")
	TokQuoted = TokenType("quoted")

	// Operators.
	TokEqual    = TokenType("equal")
	TokDot      = TokenType("dot")
	TokColon    = TokenType("colon")
	TokLParen   = TokenType("lparen")
	TokRParen   = TokenType("rparen")
	TokLBracket = TokenType("lbracket")
	TokRBracket = TokenType("rbracket")
	TokArrow    = TokenType("arrow")
	TokSlash    = TokenType("slash")

	// Keywords retained for legacy tokenizer/scanner coverage.
	TokInport  = TokenType("inport")
	TokOutport = TokenType("outport")
)

const (
	// Legacy aliases used by the experimental tokenizer implementation/tests.
	tokIllegal    = TokIllegal
	tokNewFile    = TokNewFile
	tokEOF        = TokEOF
	tokWhitespace = TokWhitespace
	tokEOL        = TokEOL
	tokComment    = TokComment
	tokIdent      = TokIdent
	tokInt        = TokInt
	tokQuotedStr  = TokQuoted
	tokEqual      = TokEqual
	tokDot        = TokDot
	tokColon      = TokColon
	tokLparen     = TokLParen
	tokRparen     = TokRParen
	tokArrow      = TokArrow
	tokSlash      = TokSlash
	tokInport     = TokInport
	tokOutport    = TokOutport
)

// Token represents a single lexeme in a File.
type Token struct {
	Type  TokenType
	File  *File
	Pos   int
	Span  Span
	Value string
}

func (t Token) String() string {
	return t.Value
}

func newCursor(file *File) Cursor {
	return Cursor{
		File:   file,
		Line:   1,
		Column: 1,
	}
}

func newToken(tokenType TokenType, cursor Cursor, end int, value string) Token {
	span := Span{Offset: cursor.Offset, Line: cursor.Line, Column: cursor.Column, End: end}
	if cursor.File != nil {
		span.File = cursor.File.Name
	}

	return Token{
		Type:  tokenType,
		File:  cursor.File,
		Pos:   cursor.Offset,
		Span:  span,
		Value: value,
	}
}

func illegalToken(cursor Cursor, end int, value string) Token {
	return newToken(TokIllegal, cursor, end, value)
}

func advanceCursor(cursor Cursor, end int) Cursor {
	if cursor.File == nil {
		cursor.Offset = end
		return cursor
	}

	if end < cursor.Offset {
		end = cursor.Offset
	}
	if end > len(cursor.File.Data) {
		end = len(cursor.File.Data)
	}

	line := cursor.Line
	column := cursor.Column
	data := cursor.File.Data

	for i := cursor.Offset; i < end; i++ {
		switch data[i] {
		case '\r':
			if i+1 < end && data[i+1] == '\n' {
				i++
			}
			line++
			column = 1
		case '\n':
			line++
			column = 1
		default:
			column++
		}
	}

	return Cursor{
		File:   cursor.File,
		Offset: end,
		Line:   line,
		Column: column,
	}
}

func isIdentStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
