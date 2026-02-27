package compiler

import (
	"strings"
)

// TokenKind classifies a raw line.
type TokenKind int

const (
	TokDialog  TokenKind = iota // [Speaker] ... "text"
	TokSystem                   // @command: value --opt val
	TokChoice                   // >> prompt
	TokOption                   // - text -> Label  (child of choice)
	TokLabel                    // # LabelName
	TokBlank                    // empty / comment line
)

// Token is the output unit of the Lexer.
type Token struct {
	Kind TokenKind
	Raw  string // trimmed source line
	Line int    // 1-based line number for error messages
}

// Tokenize scans src line-by-line and returns the token stream.
// It never returns an error; unrecognised lines become TokBlank.
func Tokenize(src string) []Token {
	lines := strings.Split(src, "\n")
	tokens := make([]Token, 0, len(lines))

	for i, raw := range lines {
		line := i + 1
		trimmed := strings.TrimSpace(raw)

		var kind TokenKind
		switch {
		case trimmed == "" || strings.HasPrefix(trimmed, "//"):
			kind = TokBlank
		case strings.HasPrefix(trimmed, "#"):
			kind = TokLabel
		case strings.HasPrefix(trimmed, "@"):
			kind = TokSystem
		case strings.HasPrefix(trimmed, ">>"):
			kind = TokChoice
		case strings.HasPrefix(trimmed, "-") && len(trimmed) > 1 && trimmed[1] == ' ':
			kind = TokOption
		case strings.HasPrefix(trimmed, "["):
			kind = TokDialog
		default:
			kind = TokBlank
		}

		tokens = append(tokens, Token{Kind: kind, Raw: trimmed, Line: line})
	}

	return tokens
}
