package components

import (
	"math"

	"github.com/sywesk/audiomix/pkg/audiograph"
)

type SinGenerator struct {
	description audiograph.ComponentDescription
	lastPhase   float64
}

func NewSinGenerator() *SinGenerator {
	return &SinGenerator{
		description: audiograph.ComponentDescription{
			Inputs: []audiograph.ComponentInput{
				{
					Name:        "freq",
					Description: "frequency of the generated sinusoïd. between 0 and sampling frequency",
					Value: audiograph.Value{
						Type: audiograph.FloatValueType,
					},
				},
				{
					Name:        "gain",
					Description: "controls the amplitude of the generated sinusoïd. min 0",
					Value: audiograph.Value{
						Type: audiograph.FloatValueType,
					},
				},
				{
					Name:        "offset",
					Description: "controls the offset of the generated sinusoïd.",
					Value: audiograph.Value{
						Type: audiograph.FloatValueType,
					},
				},
			},
			Outputs: []audiograph.ComponentOutput{
				{
					Name:        "sinusoid",
					Description: "sinusoid curve",
					Value: audiograph.Value{
						Type: audiograph.FloatValueType,
					},
				},
			},
		},
	}
}

func (s *SinGenerator) GetDescription() *audiograph.ComponentDescription {
	return &s.description
}

func (s *SinGenerator) Execute(ctx audiograph.ExecutionContext) error {
	freq := s.description.Inputs[0].Value.Float
	gain := s.description.Inputs[1].Value.Float
	offset := s.description.Inputs[2].Value.Float

	phaseIncrement := 2 * math.Pi * freq / float64(ctx.SamplingFrequency)
	newPhase := math.Mod(s.lastPhase+phaseIncrement, 2*math.Pi)

	s.description.Outputs[0].Value.Float = math.Sin(newPhase)*gain + offset
	s.lastPhase = newPhase

	return nil
}
