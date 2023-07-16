package main

import (
	"time"

	"github.com/hajimehoshi/oto/v2"
	"github.com/sywesk/audiomix/pkg/audiograph"
	"github.com/sywesk/audiomix/pkg/audiograph/components"
)

const (
	SAMPLE_RATE = 48000
)

func main() {
	otoCtx, ctxReady, err := oto.NewContext(SAMPLE_RATE, 2, oto.FormatSignedInt16LE)
	if err != nil {
		panic("failed to init oto: " + err.Error())
	}

	<-ctxReady

	graph := audiograph.New(SAMPLE_RATE)

	freqID := graph.AddComponent(components.NewFloatParam(650.0))
	gainID := graph.AddComponent(components.NewFloatParam(1.0))
	offsetID := graph.AddComponent(components.NewFloatParam(0.0))
	sinID := graph.AddComponent(components.NewSinGenerator())
	f2sID := graph.AddComponent(components.NewFloatToSample())

	graph.MustAddCable(freqID, "float", sinID, "freq")
	graph.MustAddCable(gainID, "float", sinID, "gain")
	graph.MustAddCable(offsetID, "float", sinID, "offset")
	graph.MustAddCable(sinID, "sinusoid", f2sID, "float")

	graph.SetOutput(f2sID, "sample")

	player := otoCtx.NewPlayer(graph)
	player.Play()

	time.Sleep(5 * time.Second)

	player.Close()
}
