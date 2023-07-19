package ddl

import (
	"fmt"
	"os"

	"github.com/sywesk/audiomix/pkg/audiograph"
)

func LoadFile(path string) (*audiograph.AudioGraph, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	tokenizer := newLexer(file)
	parser := newParser(tokenizer)
	interpreter := newInterpreter(parser)

	err = interpreter.BuildGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	return interpreter.GetGraph(), nil
}
