package audiograph

import (
	"fmt"
	"sync"
)

var (
	ErrInputAlreadyUsed     = fmt.Errorf("input already used")
	ErrUnknownCable         = fmt.Errorf("unknown cable")
	ErrUnknownComponent     = fmt.Errorf("unknown component")
	ErrUnknownComponentPort = fmt.Errorf("unknown component port")
)

type PortLocation int

const (
	InputPortLocation  PortLocation = 1
	OutputPortLocation PortLocation = 2
)

type PortAddress struct {
	ComponentID ComponentID
	ConnectorID uint
}

func (p PortAddress) String() string {
	return fmt.Sprintf("(component: %d, conn: %d)", p.ComponentID, p.ConnectorID)
}

type CableID uint64
type ComponentID uint64

type Cable struct {
	Source      PortAddress
	Destination PortAddress
}

type audioGraphComponent struct {
	component   Component
	description *ComponentDescription
	deleted     bool
	inputNames  map[string]uint
	outputNames map[string]uint
}

type audioGraphCable struct {
	cable   Cable
	deleted bool
}

type AudioGraph struct {
	mutex sync.RWMutex

	samplingFrequency uint32
	output            PortAddress
	outputSet         bool

	components       []audioGraphComponent
	cables           []audioGraphCable
	freeComponentIDs []ComponentID
	freeCableIDs     []CableID

	// cableSourceIndex allows to know which cable is connected to an output.
	// There may be multiple cables originating from a single output.
	cableSourceIndex map[PortAddress][]CableID

	// cableDestIndex allows to know which cable is connected to an input.
	// There can be only **one** cable connected to a single input.
	cableDestIndex map[PortAddress]CableID
}

func New(samplingFrequency uint32) *AudioGraph {
	return &AudioGraph{
		cableSourceIndex:  map[PortAddress][]CableID{},
		cableDestIndex:    map[PortAddress]CableID{},
		samplingFrequency: samplingFrequency,
	}
}

func (a *AudioGraph) AddComponent(component Component) ComponentID {
	inputNames := map[string]uint{}
	for id, input := range component.GetDescription().Inputs {
		inputNames[input.Name] = uint(id)
	}

	outputNames := map[string]uint{}
	for id, output := range component.GetDescription().Outputs {
		outputNames[output.Name] = uint(id)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	id := a.getNextComponentID()

	a.components[id] = audioGraphComponent{
		component:   component,
		description: component.GetDescription(),
		deleted:     false,
		inputNames:  inputNames,
		outputNames: outputNames,
	}

	return id
}

func (a *AudioGraph) SetOutput(outputComponentID ComponentID, outputPort string) error {
	portAddr, err := a.ResolvePortAddr(outputComponentID, outputPort, OutputPortLocation)
	if err != nil {
		return fmt.Errorf("failed to resolve port addr: %w", err)
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	a.output = portAddr
	a.outputSet = true

	return nil
}

func (a *AudioGraph) MustResolvePortAddr(componentID ComponentID, portName string, location PortLocation) PortAddress {
	pa, err := a.ResolvePortAddr(componentID, portName, location)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve port addr (%d->%s, %d): %v", componentID, portName, location, err))
	}

	return pa
}

func (a *AudioGraph) ResolvePortAddr(componentID ComponentID, portName string, location PortLocation) (PortAddress, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if componentID >= ComponentID(len(a.components)) || a.components[componentID].deleted {
		return PortAddress{}, ErrUnknownComponent
	}

	container := a.components[componentID].inputNames
	if location == OutputPortLocation {
		container = a.components[componentID].outputNames
	}

	connectorID, ok := container[portName]
	if !ok {
		return PortAddress{}, ErrUnknownComponentPort
	}

	return PortAddress{
		ComponentID: componentID,
		ConnectorID: connectorID,
	}, nil
}

func (a *AudioGraph) getNextComponentID() ComponentID {
	if len(a.freeComponentIDs) > 0 {
		nextID := a.freeComponentIDs[0]
		a.freeComponentIDs = a.freeComponentIDs[1:]

		return nextID
	}

	nextID := len(a.components)
	a.components = append(a.components, audioGraphComponent{deleted: true})

	return ComponentID(nextID)
}

func (a *AudioGraph) DeleteComponent(id ComponentID) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if id >= ComponentID(len(a.components)) || a.components[id].deleted {
		return ErrUnknownComponent
	}

	component := a.components[id]

	// Remove all input cables
	for portID := range component.description.Inputs {
		cableID, ok := a.cableDestIndex[PortAddress{
			ComponentID: id,
			ConnectorID: uint(portID),
		}]
		if !ok {
			// No cable to remove
			continue
		}

		a.deleteCable(cableID)
	}

	// Remove all output cables
	for portID := range component.description.Outputs {
		cableIDs, ok := a.cableSourceIndex[PortAddress{
			ComponentID: id,
			ConnectorID: uint(portID),
		}]
		if !ok {
			// No cable to remove
			continue
		}

		for _, cableID := range cableIDs {
			a.deleteCable(cableID)
		}
	}

	// Remove the component itself
	a.components[id].deleted = true
	a.freeComponentIDs = append(a.freeComponentIDs, id)

	return nil
}

func (a *AudioGraph) MustAddCable(sourceComponentID ComponentID, sourcePort string, destComponentID ComponentID, destPort string) CableID {
	cableID, err := a.AddCable(sourceComponentID, sourcePort, destComponentID, destPort)
	if err != nil {
		panic(fmt.Sprintf("failed to add cable: %w", err))
	}

	return cableID
}

func (a *AudioGraph) AddCable(sourceComponentID ComponentID, sourcePort string, destComponentID ComponentID, destPort string) (CableID, error) {
	sourcePortAddr, err := a.ResolvePortAddr(sourceComponentID, sourcePort, OutputPortLocation)
	if err != nil {
		return CableID(0), fmt.Errorf("failed to resolve source port addr: %w", err)
	}

	destPortAddr, err := a.ResolvePortAddr(destComponentID, destPort, InputPortLocation)
	if err != nil {
		return CableID(0), fmt.Errorf("failed to resolve dest port addr: %w", err)
	}

	return a.addCable(Cable{
		Source:      sourcePortAddr,
		Destination: destPortAddr,
	})
}

func (a *AudioGraph) addCable(cable Cable) (CableID, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, ok := a.cableDestIndex[cable.Destination]; ok {
		return 0, fmt.Errorf("input %s cannot be used: %w", cable.Destination.String(), ErrInputAlreadyUsed)
	}

	id := a.getNextCableID()

	a.cables[id] = audioGraphCable{
		cable:   cable,
		deleted: false,
	}
	a.cableDestIndex[cable.Destination] = id
	a.cableSourceIndex[cable.Source] = append(a.cableSourceIndex[cable.Source], id)

	return id, nil
}

func (a *AudioGraph) getNextCableID() CableID {
	if len(a.freeCableIDs) > 0 {
		nextID := a.freeCableIDs[0]
		a.freeCableIDs = a.freeCableIDs[1:]

		return nextID
	}

	nextID := len(a.cables)
	a.cables = append(a.cables, audioGraphCable{deleted: true})

	return CableID(nextID)
}

func (a *AudioGraph) DeleteCable(id CableID) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if id >= CableID(len(a.cables)) || a.cables[id].deleted {
		return ErrUnknownCable
	}

	a.deleteCable(id)
	return nil
}

func (a *AudioGraph) deleteCable(id CableID) {
	cable := a.cables[id].cable

	// 1. Remove destination from the index
	delete(a.cableDestIndex, cable.Destination)

	// 2. Remove the source from the index
	sourceCables := filterOutCableID(a.cableSourceIndex[cable.Source], id)
	if len(sourceCables) == 0 {
		delete(a.cableSourceIndex, cable.Source)
	} else {
		a.cableSourceIndex[cable.Source] = sourceCables
	}

	// 3. Remove the actual cable
	a.cables[id].deleted = true
	a.freeCableIDs = append(a.freeCableIDs, id)
}

func (a *AudioGraph) iterate() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 1. Follow cables to copy values
	for _, cable := range a.cables {
		srcAddr := cable.cable.Source
		dstAddr := cable.cable.Destination

		srcDesc := a.components[srcAddr.ComponentID].description
		dstDesc := a.components[dstAddr.ComponentID].description

		srcDesc.Outputs[srcAddr.ConnectorID].Value.CopyTo(&dstDesc.Inputs[dstAddr.ConnectorID].Value)
	}

	// 2. Execute all components
	ctx := ExecutionContext{
		SamplingFrequency: a.samplingFrequency,
	}

	for id, component := range a.components {
		err := component.component.Execute(ctx)
		if err != nil {
			return fmt.Errorf("failed to execute component %d: %w", id, err)
		}
	}

	return nil
}

func filterOutCableID(ids []CableID, id CableID) []CableID {
	if len(ids) == 0 {
		return nil
	}

	for i, cableId := range ids {
		if cableId != id {
			continue
		}

		ids = append(ids[:i], ids[i+1:]...)
		break
	}

	return ids
}

func (a *AudioGraph) Read(p []byte) (n int, err error) {
	// 4 bytes, 2 for the left channel, 2 for the Right
	sampleSize := 4

	if len(p)%sampleSize != 0 {
		panic("audioProcuder.Read: p is not a multiple of sampleSize")
	}

	requestedSamples := len(p) / sampleSize

	if requestedSamples > 500 {
		requestedSamples = 500
	}

	for i := 0; i < requestedSamples; i++ {
		err := a.iterate()
		if err != nil {
			return 0, fmt.Errorf("failed to compute iteration: %w", err)
		}

		left := int16(0)
		right := int16(0)

		if !a.outputSet {
			left = 0
			right = 0
		} else {
			component := a.components[a.output.ComponentID]
			value := component.description.Outputs[a.output.ConnectorID].Value

			left = int16(value.Sample.Left)
			right = int16(value.Sample.Right)
		}

		p[i*sampleSize+0] = byte(left >> 0)
		p[i*sampleSize+1] = byte(left >> 8)
		p[i*sampleSize+2] = byte(right >> 0)
		p[i*sampleSize+3] = byte(right >> 8)
	}

	return requestedSamples * sampleSize, nil
}
