package dac

import (
	"fmt"
	"testing"
)

type fakeObjectAdapter struct {
	objects map[string]Object
}

func (wr *fakeObjectAdapter) Write(obj Object) error {
	wr.objects[string(obj.ID[:])] = obj
	return nil
}

func (wr *fakeObjectAdapter) Read(id []byte) (Object, error) {
	var obj, found = wr.objects[string(id)]
	if !found {
		return Object{}, fmt.Errorf("Cannot find %#x", id)
	}

	return obj, nil
}

func TestNodeInsertion(t *testing.T) {
	// Arrange
	var fakeAdapter = new(fakeObjectAdapter)
	fakeAdapter.objects = map[string]Object{}

	var graph, err = NewDACGraph(fakeAdapter, nil)
	if err != nil {
		t.Errorf("Could not initialize the graph correctly: %s", err.Error())
		t.FailNow()
	}

	rootObj, _ := graph.AppendNode([]byte("HalloWorld"), []ObjectID{}...)

	// Act
	obj, err := graph.AppendNode([]byte("HalloWorld"), rootObj.ID)

	//Assert
	if err != nil {
		t.FailNow()
	}

	if obj == nil {
		t.Errorf("Appended node was nil")
		t.FailNow()
	}
	if len(obj.PredecessorIDs) != 1 {
		t.Errorf("Appended nodes predecessor collection is of unexpected length %v but should be %v", len(obj.PredecessorIDs), 1)
		t.FailNow()
	}

}
