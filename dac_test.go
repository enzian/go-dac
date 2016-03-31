package dac

import (
	"fmt"
	"testing"
)

type fakeObjectAdapter struct {
	objects map[string]Object
	refs    map[string]Reference
}

func (wr *fakeObjectAdapter) WriteObject(obj Object) error {
	wr.objects[string(obj.ID[:])] = obj
	return nil
}

func (wr *fakeObjectAdapter) ReadObject(id []byte) (Object, error) {
	var obj, found = wr.objects[string(id)]
	if !found {
		return Object{}, fmt.Errorf("Cannot find %#x", id)
	}

	return obj, nil
}

func (wr *fakeObjectAdapter) ReadReference(id string) (Reference, bool, error) {
	var ref, found = wr.refs[id]
	if !found {
		return Reference{Name: id}, found, nil
	}

	return ref, true, nil
}

func (wr *fakeObjectAdapter) WriteReference(ref Reference) error {
	wr.refs[ref.Name] = ref
	return nil
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

	rootObj, _ := graph.AppendNode([]byte("HelloWorld"), []ObjectID{}...)

	// Act
	obj, err := graph.AppendNode([]byte("HelloWorld"), rootObj.ID)

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

func TestNodeInsertion_WithReferences(t *testing.T) {
	// Arrange
	var fakeAdapter = new(fakeObjectAdapter)
	fakeAdapter.objects = map[string]Object{}
	fakeAdapter.refs = map[string]Reference{}
	var refName = "master"
	var graph, _ = NewDACGraph(fakeAdapter, fakeAdapter)

	var newObj, err = graph.AppendNodeToRef([]byte("HelloWorld"), refName)
	if err != nil {
		t.Errorf("Failed to insert a new node to reference %s: %s", refName, err.Error())
		t.FailNow()
	}

	var masterRef, found, _ = fakeAdapter.ReadReference(refName)
	if !found {
		t.Errorf("Appending a node to the (non-existent) reference %s did not create it!", refName)
		t.FailNow()
	}
	if masterRef.TargetID != newObj.ID {
		t.Errorf("Appending a node to reference %s, did not move the reference correctly. Expected position: \n %#x \n but was at: \n %#x", refName, newObj.ID, masterRef.TargetID)
		t.FailNow()
	}
}
