package audiograph

type SampleType int16

type Sample struct {
	Left  SampleType
	Right SampleType
}

type ValueType int

const (
	IntegerValueType ValueType = 1
	FloatValueType   ValueType = 2
	SampleValueType  ValueType = 3
	BoolValueType    ValueType = 4
)

type Value struct {
	Type    ValueType
	Integer int64
	Float   float64
	Sample  Sample
	Bool    bool
}

func (v Value) CopyTo(dest *Value) {
	dest.Type = v.Type
	dest.Integer = v.Integer
	dest.Float = v.Float
	dest.Sample = v.Sample
	dest.Bool = v.Bool
}

type ComponentInput struct {
	Name        string
	Description string
	Value       Value
}

type ComponentOutput struct {
	Name        string
	Description string
	Value       Value
}

type ComponentDescription struct {
	Inputs  []ComponentInput
	Outputs []ComponentOutput
}

type ExecutionContext struct {
	SamplingFrequency uint32
}

type Component interface {
	GetDescription() *ComponentDescription
	Execute(ExecutionContext) error
}