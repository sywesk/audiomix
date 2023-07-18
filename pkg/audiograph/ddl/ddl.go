package ddl

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"os"

	"github.com/sywesk/audiomix/pkg/audiograph"
)

func LoadFile(path string) (*audiograph.AudioGraph, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	tokenizer := NewTokenizer(file)

	parser := NewParser(tokenizer)

	for {
		stmt, err := parser.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next statement: %w", err)
		}

		spew.Dump(stmt)
	}
}
