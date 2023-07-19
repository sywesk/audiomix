package main

import (
	"fmt"
	"github.com/hajimehoshi/oto/v2"
	"github.com/sywesk/audiomix/pkg/audiograph/ddl"
	"time"
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

	graph, err := ddl.LoadFile("./examples/sin_sin.audiograph")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	player := otoCtx.NewPlayer(graph)
	player.Play()

	time.Sleep(5 * time.Second)

	player.Close()
}
