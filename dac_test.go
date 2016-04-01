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

func TestAttachReference(t *testing.T) {
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

	// Act
	ref, err := graph.Reference(newObj.ID, "master")

	// Assert
	if err != nil {
		t.Errorf("Creation of reference %s lead to an unexpected error: %s", refName, err.Error())
		t.FailNow()
	}
	if ref.TargetID != newObj.ID {
		t.Errorf("Reference %s was successfully created but did not point to the correct object id.\n Expected:\n\t %#x \n but found: \n\t %#x", refName, newObj.ID, ref.TargetID)
		t.FailNow()
	}
}

/*
This test builds a more complex DGA and finds the LCA (lowest common ancestor)
of two given references (source, target):

*   f2fdb6b - objE  (ref:source)
| * 998b1c4 - objD  (ref:target)
| * 96a0128 - objC
| * 5deee5d - objB
|/
*   2d9287b4 - objA  (should be the LCA)


*/
func TestFindLCA_Simple(t *testing.T) {
	// Arrange
	var fakeAdapter = new(fakeObjectAdapter)
	fakeAdapter.objects = map[string]Object{}
	fakeAdapter.refs = map[string]Reference{}
	var srcRef, targetRef = "source", "target"
	var graph, _ = NewDACGraph(fakeAdapter, fakeAdapter)

	objA, err := graph.AppendNode([]byte("HelloWorld"), ObjectID{})

	_, err = graph.Reference(objA.ID, targetRef)
	_, err = graph.Reference(objA.ID, srcRef)

	_, err = graph.AppendNodeToRef([]byte("HelloWorld"), targetRef)
	_, err = graph.AppendNodeToRef([]byte("HelloWorld"), targetRef)
	_, err = graph.AppendNodeToRef([]byte("HelloWorld"), targetRef)
	_, err = graph.AppendNodeToRef([]byte("HelloWorld"), srcRef)

	if err != nil {
		t.Errorf("Failed to insert a new node or reference: %s", err.Error())
		t.FailNow()
	}

	// Act
	lcaObject, err := graph.FindLowestCommonAncestor(targetRef, srcRef)

	// Assert
	if err != nil {
		t.Errorf("Failed to find LCA: %s", err.Error())
		t.FailNow()
	}
	if lcaObject == nil {
		t.Errorf("Failed to find LCA because it was Nil.")
		t.FailNow()
	}
	if lcaObject.ID != objA.ID {
		t.Errorf("Failed to find correct LCA. \n Expected: \n\t %#x \n but found: \n\t %#x \n", objA.ID, lcaObject)
		t.FailNow()
	}
}

/*
This test builds a DGA that will produce two LCAs for source and target:
*/
func TestFindLCA_Multiple(t *testing.T) {
	// Arrange
	var fakeAdapter = new(fakeObjectAdapter)
	fakeAdapter.objects = map[string]Object{}
	fakeAdapter.refs = map[string]Reference{}
	var srcRef, tarRef = "source", "target"
	var graph, _ = NewDACGraph(fakeAdapter, fakeAdapter)

	lcaA, _ := graph.AppendNode([]byte("HelloWorld"), ObjectID{})
	lcaB, _ := graph.AppendNode([]byte("HelloWorld"), lcaA.ID)

	startA, _ := graph.AppendNode([]byte("HelloWorld"), lcaA.ID, lcaB.ID)
	startB, _ := graph.AppendNode([]byte("HelloWorld"), lcaB.ID, lcaA.ID)

	var sourceRef, _ = graph.Reference(startA.ID, srcRef)
	var targetRef, _ = graph.Reference(startB.ID, tarRef)

	// Act
	var lcaObject, err = graph.FindLowestCommonAncestor(sourceRef.Name, targetRef.Name)

	// Assert
	if err != nil {
		t.Errorf("Encountered unexpected error while looking for the lca: %s", err.Error())
		t.FailNow()
	} else if lcaObject.ID != lcaA.ID && lcaObject.ID != lcaB.ID {
		t.Errorf("Looking for lca in graph with multiple lca candidates returned none of the two possible correct lcas")
		t.FailNow()
	}
}

/*
This test builds a DGA that will be iterated from multiple points to find the LCA
*/
func TestFindLCA_MultipleSearchRoots(t *testing.T) {
	// Arrange
	var fakeAdapter = new(fakeObjectAdapter)
	fakeAdapter.objects = map[string]Object{}
	fakeAdapter.refs = map[string]Reference{}
	var bRef, cRef, dRef = "B", "C", "D"
	var graph, _ = NewDACGraph(fakeAdapter, fakeAdapter)

	root, _ := graph.AppendNode([]byte("Root Object"), ObjectID{})

	objB, _ := graph.AppendNode([]byte("Frist Object on B"), root.ID)
	objB, _ = graph.AppendNode([]byte("Second Object on B"), objB.ID)
	graph.Reference(objB.ID, bRef)

	objC, _ := graph.AppendNode([]byte("Frist Object on C"), root.ID)
	objC, _ = graph.AppendNode([]byte("Second Object on C"), objC.ID)
	graph.Reference(objC.ID, cRef)

	objD, _ := graph.AppendNode([]byte("Frist Object on D"), root.ID)
	objD, _ = graph.AppendNode([]byte("Second Object on D"), objD.ID)
	graph.Reference(objD.ID, dRef)

	// Act
	var lcaObject, err = graph.FindLowestCommonAncestor(bRef, cRef, dRef)

	// Assert
	if err != nil {
		t.Errorf("Encountered unexpected error while looking for the lca: %s", err.Error())
		t.FailNow()
	} else if lcaObject.ID != root.ID {
		t.Errorf("Looking up the LCA failed with multiple search roots. Expected %#x but found %#x", root.ID[:4], lcaObject.ID[:4])
		t.FailNow()
	}
}
