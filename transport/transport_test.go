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
	var from = dac.IdFromString("4e0ec4f2bddac0eefa18b8560edf158be96fcaccf6644f1d6a41632c8e31e3cf4e0ec4f2bddac0eefa18b8560edf158be96fcaccf6644f1d6a41632c8e31e3cf")
	var via = dac.IdFromString("ec4f2bdd8be96fcaccf658be96fcaccf644f1d6a41632c8e31e3cf4e06644f1d6aedf14e0ec4f2bddac0eefa18b8560edf1541632c8e31e3cfac0eefa18b8560")
	var viaContent = []byte("Content VIA")
	var to = dac.IdFromString("644f1d6a41632c8e31e3cf4e0ec4f2bdd8be96fcaccf658be96fcaccf6644f1d6a41632c8e31e3cfac0eefa18b8560edf14e0ec4f2bddac0eefa18b8560edf15")
	var toContent = []byte("Content TO")
	var name = "refs/heads/master"
	var command = hex.EncodeToString(from[:])
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

	var ret map[dac.ObjectID]dac.Object

	var unlock = make(chan bool, 1)

	var objectSupplier = func() <-chan dac.Object {
		var objChan = make(chan dac.Object, 2)
		objChan <- dac.Object{ID: via, Content: viaContent, PredecessorIDs: []dac.ObjectID{from}}
		objChan <- dac.Object{ID: to, Content: toContent, PredecessorIDs: []dac.ObjectID{via}}
		close(objChan)
		return objChan
	}

	go func() {
		ret = ReceiveObjects(graph, sendChan, rcvChan, objectSupplier)
		unlock <- true
	}()

	<-unlock

	// Assert

	var resultSize = len(ret)
	assert.Equal(t, 2, resultSize, "Expected to found two objects in the update package")

	var _, exists = ret[via]
	assert.Equal(t, true, exists, "Expected the intermediary object 'via' (%#x) to be in the update package.", via[:4])

	_, exists = ret[to]
	assert.Equal(t, true, exists, "Expected the object 'to' (%#x) to be in the update package", to[:4])
}
