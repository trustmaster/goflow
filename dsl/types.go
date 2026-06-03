package dsl

import (
	"unicode"
	"unicode/utf8"
)

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
)

const (
	keywordINPORT  = "INPORT"
	keywordOUTPORT = "OUTPORT"
)

// Token represents a single lexeme in a File.
type Token struct {
	Type  TokenType
	File  *File  // reference to the source file (same as Span.File for convenience)
	Pos   int    // byte offset in file (same as Span.Offset)
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

	for offset := cursor.Offset; offset < end; {
		r, size := utf8.DecodeRune(data[offset:])
		if r == '\n' {
			line++
			column = 1
		} else if r == '\r' {
			if offset+size < end {
				nextR, _ := utf8.DecodeRune(data[offset+size:])
				if nextR == '\n' {
					offset += size
				}
			}
			line++
			column = 1
		} else {
			column++
		}
		offset += size
	}

	return Cursor{
		File:   cursor.File,
		Offset: end,
		Line:   line,
		Column: column,
	}
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || unicode.IsDigit(r)
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
