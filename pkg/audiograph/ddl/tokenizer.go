package ddl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type TokenType string

const (
	UnknownToken            TokenType = ""
	AtToken                 TokenType = "@"
	OpeningParenthesisToken TokenType = "("
	ClosingParenthesisToken TokenType = ")"
	EqualToken              TokenType = "="
	ConnectToken            TokenType = "->"
	IdentifierToken         TokenType = "id"
	NumberToken             TokenType = "n"
	ColonToken              TokenType = ":"
	ReturnToken             TokenType = "r"
)

var (
	specialSymbols = map[string]TokenType{
		"@":  AtToken,
		"(":  OpeningParenthesisToken,
		")":  ClosingParenthesisToken,
		"=":  EqualToken,
		"->": ConnectToken,
		":":  ColonToken,
		"\n": ReturnToken,
	}
)

type Token struct {
	Value string
	Type  TokenType
	Line  int
	Col   int
}

func (t Token) String() string {
	return fmt.Sprintf("token:[value:'%s', type:'%s', line:%d, col:%d]",
		t.Value,
		string(t.Type),
		t.Line,
		t.Col)
}

type Tokenizer struct {
	reader     *bufio.Reader
	readBuffer []byte
	line       int
	col        int
}

func NewTokenizer(reader io.Reader) *Tokenizer {
	return &Tokenizer{
		reader:     bufio.NewReader(reader),
		readBuffer: []byte{0},
		line:       1,
		col:        1,
	}
}

func (t *Tokenizer) Next() (Token, error) {
	token := Token{
		Value: "",
		Type:  UnknownToken,
		Line:  t.line,
		Col:   t.col,
	}

	for {
		r, _, err := t.reader.ReadRune()
		if err != nil {
			return token, err
		}

		// Update col & line numbers
		if r == '\n' {
			t.line++
			t.col = 1
		} else {
			t.col++
		}

		// Skip initial spaces
		if token.Type == UnknownToken && r != '\n' && unicode.IsSpace(r) {
			continue
		}

		// If we're in a token and we reach an end of line, unread the line return
		// for it to be sent as a token at the next Next() call.
		if token.Type != UnknownToken && r == '\n' {
			_ = t.reader.UnreadRune()
			break
		}

		// A "white space" during a token is considered the end of that token
		if token.Type != UnknownToken && r != '\n' && unicode.IsSpace(r) {
			break
		}

		if token.Type == IdentifierToken {
			if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_' || r == '-' {
				token.Value += string(r)
				continue
			}

			// Unknown rune for this token, unread and return
			_ = t.reader.UnreadRune()
			break
		}

		if token.Type == NumberToken {
			if unicode.IsDigit(r) || r == '.' {
				// a number cannot contain more than 1 '.'
				if r == '.' && strings.IndexRune(token.Value, r) >= 0 {
					return token, fmt.Errorf("two many '.' in a number: %s", token.String())
				}

				token.Value += string(r)
				continue
			}

			// Unknown rune for this token, unread and return
			_ = t.reader.UnreadRune()
			break
		}

		token.Value += string(r)

		// Determine the token type
		if token.Type == UnknownToken {
			// maybe it's the beginning of a connect symbol, loop another time
			if len(token.Value) == 1 && r == '-' {
				continue
			}

			// look for a special symbol. this will also catch connect symbols.
			tokenType, ok := specialSymbols[token.Value]
			if ok {
				token.Type = tokenType
				break
			}

			// we don't want to catch "-" + letter
			if len(token.Value) == 1 && (unicode.IsLetter(r) || r == '_') {
				token.Type = IdentifierToken
				continue
			}

			// there could be a "-" before, which would act as a sign
			if unicode.IsDigit(r) {
				token.Type = NumberToken
				continue
			}

			return token, fmt.Errorf("failed to determine token type: %s", token.String())
		}

		return token, fmt.Errorf("unexpected token type: %s", token.String())
	}

	return token, nil
}
