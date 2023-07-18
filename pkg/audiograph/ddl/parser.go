package ddl

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrUnexpectedToken = fmt.Errorf("unexpected token")
	ErrSyntaxError     = fmt.Errorf("syntax error")
)

type ValueType int

const (
	StringValueType ValueType = 1
	NumberValueType ValueType = 2
)

type Value struct {
	Type   ValueType
	String string
	Number float64
}

type Connector struct {
	VariableName  string
	ConnectorName string
}

type StatementType int

const (
	ParameterStatementType       StatementType = 1
	CreateComponentStatementType StatementType = 2
	ConnectStatementType         StatementType = 3
)

type Statement interface {
	Type() StatementType
}

type ParameterStatement struct {
	Name  string
	Value Value
}

func (p ParameterStatement) Type() StatementType {
	return ParameterStatementType
}

type CreateComponentStatement struct {
	VariableName  string
	ComponentName string
	Arguments     []Value
}

func (p CreateComponentStatement) Type() StatementType {
	return CreateComponentStatementType
}

type ConnectStatement struct {
	From Connector
	To   Connector
}

func (p ConnectStatement) Type() StatementType {
	return ConnectStatementType
}

type ILexer interface {
	Next() (Token, error)
}

type Parser struct {
	lexer ILexer
}

func NewParser(lexer ILexer) *Parser {
	return &Parser{
		lexer: lexer,
	}
}

func (p *Parser) Next() (Statement, error) {
	var tokens []Token

	for {
		token, err := p.lexer.Next()
		if err != nil {
			if errors.Is(err, io.EOF) && len(tokens) > 0 {
				break
			}

			return nil, fmt.Errorf("failed to get next token: %w", err)
		}

		if token.Type == ReturnToken {
			// Skip through the heading line feeds, and stop statement parsing at the next return
			if len(tokens) == 0 {
				continue
			} else {
				break
			}
		}

		tokens = append(tokens, token)
	}

	// Now, parse the statement tokens

	var statement Statement
	var err error

	switch tokens[0].Type {
	case AtToken:
		statement, err = p.parseParamStatement(tokens)
	case IdentifierToken:
		if len(tokens) < 1 {
			return nil, fmt.Errorf("no token after %s: %w", tokens[0].String(), ErrSyntaxError)
		}

		// TODO

	default:
		return nil, fmt.Errorf("unexpected token %s: %w", tokens[0].String(), ErrUnexpectedToken)
	}

	return statement, err
}

func (p *Parser) parseParamStatement(tokens []Token) (ParameterStatement, error) {
	// TODO
}
