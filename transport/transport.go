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
	From                           dac.ObjectID
	To                             dac.ObjectID
	findObjectWithGivenPredecessor string
}

type ObjectUnmarshaller func() <-chan dac.Object

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

func ReceiveObjects(graph *dac.Graph, recvLines <-chan ProtoLine, sends chan<- ProtoLine, objectSrc ObjectUnmarshaller) (map[dac.ObjectID]dac.Object, error) {
	RefDiscovery(graph, sends)
	var updates = readReferenceUpdates(recvLines)

	// read all objects from the line
	var objects = map[dac.ObjectID]dac.Object{}
	for obj := range objectSrc() {
		objects[obj.ID] = obj
	}

	// Check if the existing References can be fast-forwarded to the new objects as recorded in var updates
	for update := range updates {
		var canFF, err = findObjectWithGivenPredecessor(objects, update.From, update.To)
		if err != nil {
			return nil, fmt.Errorf("Cannot Fast-Forward reference %s from %#x to %#x", update.RefName, update.From[:4], update.To[:4])
		}
	}

	return objects, nil
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

func findObjectWithGivenPredecessor(pack dac.Pack, findID dac.ObjectID, startFrom dac.ObjectID) (bool, error) {
	var backlog = []dac.ObjectID{startFrom}

	for currentObjID := range backlog {
		var currentObj, inPack = pack[currentObjID]
		if !inPack {
			return false, fmt.Errorf("An object pointed to an ancestor (%#x) that is not in the pack", currentObjID[:4])
		}

		for ancestorID := range currentObj.PredecessorIDs {
			if ancestorID == findID {
				return true, nil
			}

			backlog = append([]dac.ObjectID{ancestorID}, backlog)
		}
	}
}
