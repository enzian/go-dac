package transport

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/go-msf/go-dac"
	"github.com/go-msf/go-dac/memory"
	"github.com/stretchr/testify/assert"
)

func TestReferenceDiscovery(t *testing.T) {
	// Arrange

	// Set up the graph
	var fakeAdapter = memory.NewMemoryAdapter()
	var bRef = "B"
	var graph, _ = dac.NewDACGraph(fakeAdapter, fakeAdapter)

	root, _ := graph.AppendNode([]byte("Root Object"), dac.ObjectID{})

	objB, _ := graph.AppendNode([]byte("First Object on B"), root.ID)
	objB, _ = graph.AppendNode([]byte("Second Object on B"), objB.ID)
	graph.Reference(objB.ID, bRef)

	// Setup the channels
	var sendChan = make(chan ProtoLine, 2)
	//var recvChan = make(chan ProtoLine, 10)

	// Act

	go RefDiscovery(graph, sendChan)

	// Assert
	var refMsg, more = <-sendChan

	if !more {
		t.Error("There should be one more message on the line")
	}

	assert.Equal(t, hex.EncodeToString(graph.References[0].TargetID[:]), refMsg.Command, "The advertised ID was not correct")
	assert.Equal(t, fmt.Sprintf("refs/heads/%s", graph.References[0].Name), string(refMsg.Content[:]), "The content of the reference advertisement line was not the correct reference name")

	flush, more := <-sendChan

	assert.Equal(t, PkgLineFlush.Command, flush.Command, "The line should be a pkg-flush (0000) but was not")
	assert.Equal(t, PkgLineFlush.Content, flush.Content[:], "The content of the pkg-flush should have been empty but was not.")
}

func TestSendPackage(t *testing.T) {
	// Arrange
	var from = "4e0ec4f2bddac0eefa18b8560edf158be96fcaccf6644f1d6a41632c8e31e3cf4e0ec4f2bddac0eefa18b8560edf158be96fcaccf6644f1d6a41632c8e31e3cf"
	var to = "644f1d6a41632c8e31e3cf4e0ec4f2bdd8be96fcaccf658be96fcaccf6644f1d6a41632c8e31e3cfac0eefa18b8560edf14e0ec4f2bddac0eefa18b8560edf15"
	var name = "refs/heads/master"
	var command = from
	var content = []byte(fmt.Sprintf("%s %s", to, name))

	var sendChan = make(chan ProtoLine, 2)
	var rcvChan = make(chan ProtoLine, 1)

	// Set up the graph
	var fakeAdapter = memory.NewMemoryAdapter()
	var graph, _ = dac.NewDACGraph(fakeAdapter, fakeAdapter)

	// Act
	sendChan <- ProtoLine{Command: command, Content: content}
	sendChan <- PkgLineFlush
	close(sendChan)

	var ret []RefUpdate

	var unlock = make(chan bool, 1)

	go func() {
		ret = ReceiveObjects(graph, sendChan, rcvChan)
		unlock <- true
	}()

	<-unlock

	// Assert

	assert.Equal(t, 1, len(ret), "Expected to receive 1 update line but did not")
	assert.Equal(t, from, hex.EncodeToString(ret[0].From[:]), "To points to a false ID")
	assert.Equal(t, to, hex.EncodeToString(ret[0].To[:]), "To points to a false ID")
}
