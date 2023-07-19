package ddl

import (
	"errors"
	"fmt"
	"github.com/sywesk/audiomix/pkg/audiograph"
	"github.com/sywesk/audiomix/pkg/audiograph/components"
	"io"
)

type IParser interface {
	Next() (Statement, error)
}

type interpreter struct {
	parser IParser
	graph  *audiograph.AudioGraph
	vars   map[string]audiograph.ComponentID

	outputSet            bool
	outputComponentIDSet bool
	outputComponentID    audiograph.ComponentID
	outputPortSet        bool
	outputPort           string
}

func newInterpreter(parser IParser) *interpreter {
	return &interpreter{
		parser: parser,
		graph:  audiograph.New(),
		vars:   map[string]audiograph.ComponentID{},
	}
}

func (i *interpreter) GetGraph() *audiograph.AudioGraph {
	return i.graph
}

func (i *interpreter) BuildGraph() error {
	for {
		stmt, err := i.parser.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("failed to get next statement: %w", err)
		}

		switch typedStmt := stmt.(type) {
		case *ParameterStatement:
			err = i.handleParameterStatement(typedStmt)
			if err != nil {
				return err
			}
		case *CreateComponentStatement:
			err = i.handleCreateComponent(typedStmt)
			if err != nil {
				return err
			}
		case *ConnectStatement:
			err = i.handleConnectStatement(typedStmt)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *interpreter) handleParameterStatement(stmt *ParameterStatement) error {
	switch stmt.Name {
	case "SAMPLING_FREQ":
		if stmt.Value.Type != audiograph.IntegerValueType {
			return fmt.Errorf("line %d: SAMPLING_FREQ expects an integer", stmt.Line)
		}
		i.graph.SetSamplingFrequency(uint32(stmt.Value.Integer))

	case "OUTPUT_COMPONENT":
		if i.outputComponentIDSet {
			return fmt.Errorf("line %d: OUTPUT_COMPONENT can be set only once", stmt.Line)
		}
		if stmt.Value.Type != audiograph.StringValueType {
			return fmt.Errorf("line %d: OUTPUT_COMPONENT expects a string", stmt.Line)
		}

		compID, ok := i.vars[stmt.Value.String]
		if !ok {
			return fmt.Errorf("line %d: unknown component '%s'", stmt.Line, stmt.Value.String)
		}

		i.outputComponentID = compID
		i.outputComponentIDSet = true

	case "OUTPUT_PORT":
		if i.outputPortSet {
			return fmt.Errorf("line %d: OUTPUT_PORT can be set only once", stmt.Line)
		}
		if stmt.Value.Type != audiograph.StringValueType {
			return fmt.Errorf("line %d: OUTPUT_PORT expects a string", stmt.Line)
		}

		i.outputPort = stmt.Value.String
		i.outputPortSet = true

	default:
		return fmt.Errorf("line %d: unknown parameter '%s'", stmt.Line, stmt.Name)
	}

	if !i.outputSet && i.outputPortSet && i.outputComponentIDSet {
		// do it first to avoid retrying if an error occurs during SetOutput
		i.outputSet = true

		err := i.graph.SetOutput(i.outputComponentID, i.outputPort)
		if err != nil {
			return fmt.Errorf("line %d: failed to set graph output: %w", stmt.Line, err)
		}
	}

	return nil
}

func (i *interpreter) handleCreateComponent(stmt *CreateComponentStatement) error {
	_, ok := i.vars[stmt.VariableName]
	if ok {
		return fmt.Errorf("line %d: variable '%s' already: %w", stmt.Line, stmt.VariableName, ErrSyntaxError)
	}

	comp, err := components.Instanciate(stmt.ComponentName)
	if err != nil {
		return fmt.Errorf("line %d: failed to instanciate component '%s': %w", stmt.Line, stmt.ComponentName, err)
	}

	compID := i.graph.AddComponent(comp)
	i.vars[stmt.VariableName] = compID

	for argName, argValue := range stmt.Arguments {
		err := i.graph.SetParameter(compID, argName, argValue)
		if err != nil {
			return fmt.Errorf("line %d: failed to set parameter '%s': %w", stmt.Line, argName, err)
		}
	}

	return nil
}

func (i *interpreter) handleConnectStatement(stmt *ConnectStatement) error {
	srcID, ok := i.vars[stmt.From.VariableName]
	if !ok {
		return fmt.Errorf("line %d: variable '%s' does not exists: %w", stmt.Line, stmt.From.VariableName, ErrSyntaxError)
	}

	dstID, ok := i.vars[stmt.To.VariableName]
	if !ok {
		return fmt.Errorf("line %d: variable '%s' does not exists: %w", stmt.Line, stmt.To.VariableName, ErrSyntaxError)
	}

	_, err := i.graph.AddCable(srcID, stmt.From.ConnectorName, dstID, stmt.To.ConnectorName)
	if err != nil {
		return fmt.Errorf("line %d: failed to add cable: %w", stmt.Line, err)
	}

	return nil
}
