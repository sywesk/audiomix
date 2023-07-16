package components

import "github.com/sywesk/audiomix/pkg/audiograph"

type FloatToSample struct {
	description audiograph.ComponentDescription
}

func NewFloatToSample() *FloatToSample {
	return &FloatToSample{
		description: audiograph.ComponentDescription{
			Inputs: []audiograph.ComponentInput{
				{
					Name:        "float",
					Description: "float to convert into an audio signal",
					Value: audiograph.Value{
						Type: audiograph.FloatValueType,
					},
				},
			},
			Outputs: []audiograph.ComponentOutput{
				{
					Name:        "sample",
					Description: "converted audio",
					Value: audiograph.Value{
						Type: audiograph.SampleValueType,
					},
				},
			},
		},
	}
}

func (s *FloatToSample) GetDescription() *audiograph.ComponentDescription {
	return &s.description
}

func (s *FloatToSample) Execute(ctx audiograph.ExecutionContext) error {
	value := s.description.Inputs[0].Value.Float

	// Quick clamping
	if value > 1.0 {
		value = 1.0
	} else if value < -1.0 {
		value = -1.0
	}

	s.description.Outputs[0].Value.Sample.Left = audiograph.SampleType(value * 32767)
	s.description.Outputs[0].Value.Sample.Right = audiograph.SampleType(value * 32767)

	return nil
}
