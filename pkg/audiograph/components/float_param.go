package components

import "github.com/sywesk/audiomix/pkg/audiograph"

type FloatParam struct {
	description audiograph.ComponentDescription
}

func NewFloatParam() *FloatParam {
	return &FloatParam{
		description: audiograph.ComponentDescription{
			Parameters: []audiograph.ComponentParameter{
				{
					Name:        "float",
					Description: "desired float value",
					Value: audiograph.Value{
						Type: audiograph.FloatValueType,
					},
				},
			},
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
	}
}

func (s *FloatParam) GetDescription() *audiograph.ComponentDescription {
	return &s.description
}

func (s *FloatParam) Execute(ctx audiograph.ExecutionContext) error {
	s.description.Outputs[0].Value.Float = s.description.Parameters[0].Value.Float
	return nil
}
