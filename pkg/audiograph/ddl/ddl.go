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

	tokenizer := NewTokenizer(file)

	for {
		token, err := tokenizer.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next token: %w", err)
		}

		fmt.Println(token.String())
	}
}
