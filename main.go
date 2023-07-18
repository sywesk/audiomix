package main

import (
	"fmt"
	"github.com/sywesk/audiomix/pkg/audiograph/ddl"
)

const (
	SAMPLE_RATE = 48000
)

func main() {
	/*otoCtx, ctxReady, err := oto.NewContext(SAMPLE_RATE, 2, oto.FormatSignedInt16LE)
	if err != nil {
		panic("failed to init oto: " + err.Error())
	}

	<-ctxReady*/

	_, err := ddl.LoadFile("./examples/sin_sin.audiograph")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return

	/*graph := audiograph.New(SAMPLE_RATE)

	sinFreqID := graph.AddComponent(components.NewSinGenerator())

	graph.MustAddCable(graph.AddComponent(components.NewFloatParam(2)), "float", sinFreqID, "freq")
	graph.MustAddCable(graph.AddComponent(components.NewFloatParam(500)), "float", sinFreqID, "gain")
	graph.MustAddCable(graph.AddComponent(components.NewFloatParam(600)), "float", sinFreqID, "offset")

	sinID := graph.AddComponent(components.NewSinGenerator())
	f2sID := graph.AddComponent(components.NewFloatToSample())

	graph.MustAddCable(sinFreqID, "sinusoid", sinID, "freq")
	graph.MustAddCable(graph.AddComponent(components.NewFloatParam(1.0)), "float", sinID, "gain")
	graph.MustAddCable(graph.AddComponent(components.NewFloatParam(0.0)), "float", sinID, "offset")
	graph.MustAddCable(sinID, "sinusoid", f2sID, "float")

	graph.SetOutput(f2sID, "sample")

	player := otoCtx.NewPlayer(graph)
	player.Play()

	time.Sleep(5 * time.Second)

	player.Close()*/
}
