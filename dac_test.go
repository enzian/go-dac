package dac

import (
	"fmt"
	"testing"
)

type fakeObjectAdapter struct {
	objects map[ObjectID]Object
}

func (wr *fakeObjectAdapter) Write(obj Object) error {
	wr.objects[obj.ID] = obj
	return nil
}

func (wr *fakeObjectAdapter) Read(id []byte) (Object, error) {
	var obj, found = wr.objects[id]
	if !found {
		return Object{}, fmt.Errorf("Cannot find %#x", id)
	}

	return obj, nil
}

func TestXxx(t *testing.T) {
	// Arrange
	var fakeAdapter = new(fakeObjectAdapter)
	fakeAdapter.objects = map[ObjectID]Object{}

	var graph, err = NewDACGraph(fakeAdapter, nil)

	if err != nil {
		t.FailNow()
	}

	// Act
	obj, err := graph.FindLowestCommonAncestor("master", "feature")

	obj, _ = graph.AppendNodeToPredecessor([]byte("HalloWorld"), nil)

	fmt.Printf("Make Object 1 with ID: %#x \n", obj.ID)

	var obj1, _ = graph.AppendNodeToPredecessor([]byte("HalloWorld"), []ObjectID{obj.ID})

	fmt.Printf("Make Object 2 with ID: %#x \n", obj1.ID)
	fmt.Printf("Object 2 with predecessors: %#x \n", obj1.PredecessorIDs)

	var obj2, _ = graph.AppendNodeToPredecessor([]byte("HalloWorld"), []ObjectID{obj.ID, obj1.ID})

	fmt.Printf("Make Object 3 with ID: %#x \n", obj2.ID)
	fmt.Printf("Object 3 with predecessors: %#x \n", obj2.PredecessorIDs)

	//Assert
	if err != nil {
		t.FailNow()
	}

	if obj == nil {
		t.FailNow()
	}

}
