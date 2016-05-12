package transport

import (
	"encoding/hex"
	"fmt"

	"github.com/go-msf/go-dac"
)

type ProtoLine struct {
	Command string
	Content []byte
}

var (
	PkgLineFlush = ProtoLine{Content: []byte(nil)}
	PkgLineDone  = ProtoLine{Command: "done", Content: []byte(nil)}
)

func RefDiscovery(graph *dac.Graph, sends chan<- ProtoLine) {
	for _, ref := range graph.References {
		var line = ProtoLine{
			Command: hex.EncodeToString(ref.TargetID[:]),
			Content: []byte(fmt.Sprintf("refs/heads/%s", ref.Name))}
		sends <- line
	}

	sends <- ProtoLine{}

	close(sends)
}
