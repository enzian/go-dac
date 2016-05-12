package transport

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/go-msf/go-dac"
)

type ProtoLine struct {
	Command string
	Content []byte
}

type RefUpdate struct {
	From    dac.ObjectID
	To      dac.ObjectID
	RefName string
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

func ReceiveObjects(graph *dac.Graph, recvLines <-chan ProtoLine, sends chan<- ProtoLine) []RefUpdate {
	RefDiscovery(graph, sends)
	var updates = readReferenceUpdates(recvLines)
	return updates
}

func readReferenceUpdates(recvLines <-chan ProtoLine) []RefUpdate {
	var updates = []RefUpdate{}

	for line := range recvLines {
		switch line.Command {
		case PkgLineFlush.Command:
			return updates
		default:
			var from, err = hex.DecodeString(line.Command)
			var fromID = dac.ObjectID{}

			if err != nil {
				continue
			}
			copy(fromID[:], from)

			var content = string(line.Content[:len(line.Content)])

			var segments = strings.Split(content, " ")
			to, err := hex.DecodeString(segments[0])

			if err != nil {
				continue
			}

			var toID = dac.ObjectID{}
			copy(toID[:], to)

			var name = segments[1]

			var update = RefUpdate{
				From:    fromID,
				To:      toID,
				RefName: name}

			updates = append(updates, update)
		}
	}

	return updates
}
