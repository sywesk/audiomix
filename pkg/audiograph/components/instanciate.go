package components

import (
	"fmt"
	"github.com/sywesk/audiomix/pkg/audiograph"
)

var (
	ErrUnknownComponent = fmt.Errorf("unknown component")
)

var (
	componentConstructorRegistry = map[string]func() audiograph.Component{
		"FloatParam":    func() audiograph.Component { return NewFloatParam() },
		"FloatToSample": func() audiograph.Component { return NewFloatToSample() },
		"SinGenerator":  func() audiograph.Component { return NewSinGenerator() },
	}
)

func Instanciate(componentName string) (audiograph.Component, error) {
	constructor, ok := componentConstructorRegistry[componentName]
	if !ok {
		return nil, ErrUnknownComponent
	}

	return constructor(), nil
}
