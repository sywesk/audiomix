package components

import "github.com/sywesk/audiomix/pkg/audiograph"

type FloatParam struct {
	description audiograph.ComponentDescription
	value       float64
}

func NewFloatParam(value float64) *FloatParam {
	return &FloatParam{
		description: audiograph.ComponentDescription{
			Outputs: []audiograph.ComponentOutput{
				{
					Name:        "float",
					Description: "desired float value",
					Value: audiograph.Value{
						Type: audiograph.FloatValueType,
					},
				},
			},
		},
		value: value,
	}
}

func (s *FloatParam) GetDescription() *audiograph.ComponentDescription {
	return &s.description
}

func (s *FloatParam) Execute(ctx audiograph.ExecutionContext) error {
	s.description.Outputs[0].Value.Float = s.value
	return nil
}
