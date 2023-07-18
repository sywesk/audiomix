package ddl

import (
	"fmt"
)

var (
	ErrSyntaxError = fmt.Errorf("syntax error")
)

type ValueType int

const (
	StringValueType  ValueType = 1
	FloatValueType   ValueType = 2
	BoolValueTYpe    ValueType = 3
	IntegerValueType ValueType = 4
)

type Value struct {
	Type    ValueType
	String  string
	Float   float64
	Integer int64
	Bool    bool
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
	Arguments     map[string]Value
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
	token, err := p.getFirstUsefulToken()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case AtToken:
		return p.parseParameter()
	case IdentifierToken:
		secondToken, err := p.lexer.Next()
		if err != nil {
			return nil, err
		}

		switch secondToken.Type {
		case ColonToken:
			return p.parseConnect(token)
		case EqualToken:
			return p.parseCreateComponent(token)
		default:
			return nil, fmt.Errorf("unexpected token %s: %w", token.String(), ErrSyntaxError)
		}
	default:
		return nil, fmt.Errorf("unexpected token %s: %w", token.String(), ErrSyntaxError)
	}
}

func (p *Parser) parseParameter() (Statement, error) {
	token, err := p.getTypedToken(IdentifierToken)
	if err != nil {
		return nil, err
	}
	paramName := token.Value

	valueToken, err := p.getOneOfTypedToken(IdentifierToken, NumberToken)
	if err != nil {
		return nil, err
	}

	value, err := valueToken.ToValue()
	if err != nil {
		return nil, err
	}

	return &ParameterStatement{
		Name:  paramName,
		Value: value,
	}, nil
}

// parseConnect parses connect expressions that look like:
//
//	<componentName>:<connectorName> -> <componentName>:<connectorName>
func (p *Parser) parseConnect(token1 Token) (Statement, error) {
	// Token1 is an Identifier
	tokens, err := p.getTypedTokens(IdentifierToken, ConnectToken, IdentifierToken, ColonToken, IdentifierToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get connect tokens: %w", err)
	}

	comp1Name := token1.Value
	comp1Conn := tokens[0].Value

	comp2Name := tokens[2].Value
	comp2Conn := tokens[4].Value

	return &ConnectStatement{
		From: Connector{
			VariableName:  comp1Name,
			ConnectorName: comp1Conn,
		},
		To: Connector{
			VariableName:  comp2Name,
			ConnectorName: comp2Conn,
		},
	}, nil
}

func (p *Parser) parseCreateComponent(token1 Token) (Statement, error) {
	// Token1 is an Identifier
	tokens, err := p.getTypedTokens(IdentifierToken, OpeningParenthesisToken)
	if err != nil {
		return nil, err
	}

	stmt := &CreateComponentStatement{
		VariableName:  token1.Value,
		ComponentName: tokens[0].Value,
		Arguments:     map[string]Value{},
	}

	token, err := p.getOneOfTypedToken(IdentifierToken, ClosingParenthesisToken)
	if err != nil {
		return nil, err
	}

	if token.Type == ClosingParenthesisToken {
		return stmt, nil
	}

	for {
		paramName := token.Value

		_, err = p.getTypedToken(EqualToken)
		if err != nil {
			return nil, err
		}

		valueToken, err := p.getOneOfTypedToken(IdentifierToken, NumberToken)
		if err != nil {
			return nil, err
		}

		value, err := valueToken.ToValue()
		if err != nil {
			return nil, err
		}

		stmt.Arguments[paramName] = value

		token, err = p.getOneOfTypedToken(ComaToken, ClosingParenthesisToken)
		if err != nil {
			return nil, err
		}

		if token.Type == ClosingParenthesisToken {
			break
		}
	}

	return stmt, nil
}

func (p *Parser) getTypedTokens(ts ...TokenType) ([]Token, error) {
	var tokens []Token

	for _, t := range ts {
		token, err := p.getTypedToken(t)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// getTypedToken gets the next token and ensures that it has the right type before returning it
func (p *Parser) getTypedToken(t TokenType) (Token, error) {
	token, err := p.lexer.Next()
	if err != nil {
		return Token{}, err
	}

	if token.Type != t {
		return Token{}, fmt.Errorf("expected token '%s' but got '%s': %w", string(t), string(token.Type), ErrSyntaxError)
	}

	return token, nil
}

func (p *Parser) getOneOfTypedToken(ts ...TokenType) (Token, error) {
	token, err := p.lexer.Next()
	if err != nil {
		return Token{}, err
	}

	// Check if the token has one of the provided types
	for _, t := range ts {
		if token.Type == t {
			return token, nil
		}
	}

	return Token{}, fmt.Errorf("unexpected token type '%s': %w", string(token.Type), ErrSyntaxError)
}

func (p *Parser) getFirstUsefulToken() (Token, error) {
	for {
		token, err := p.lexer.Next()
		if err != nil {
			return Token{}, fmt.Errorf("failed to get next token: %w", err)
		}

		if token.Type == ReturnToken {
			continue
		}

		return token, nil
	}
}
