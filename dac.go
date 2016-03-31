package dac

import (
	"fmt"
	"math"

	"github.com/go-msf/go-dac/tree"

	"golang.org/x/crypto/sha3"
)

// Object represents a node in the directed acyclic graph
type Object struct {
	ID             ObjectID
	Content        []byte
	PredecessorIDs []ObjectID
}

// An ObjectID represents an objects id expressend in a hash
type ObjectID [64]byte

// EmptyID describes an empty ID for comparison
var EmptyID = ObjectID{}

// A Reference points to an object in the graph
type Reference struct {
	TargetID ObjectID
	Name     string
}

// The Graph represents a number of nodes and references pointing to these nodes
type Graph struct {
	References []Reference
	ObjectAdapter
	ReferenceAdapter
}

// ObjectReader reads objects
type ObjectReader interface {
	ReadObject(id []byte) (Object, error)
}

// ObjectWriter persist new/changed objects
type ObjectWriter interface {
	WriteObject(obj Object) error
}

// ObjectAdapter extends read and write capabilities for objects to the graph
type ObjectAdapter interface {
	ObjectReader
	ObjectWriter
}

// ReferenceReader reads references
type ReferenceReader interface {
	ReadReference(name string) (Reference, bool, error)
}

// ReferenceWriter persist new/changed references
type ReferenceWriter interface {
	WriteReference(obj Reference) error
}

// ReferenceAdapter extends read and write capabilities for objects to the graph
type ReferenceAdapter interface {
	ReferenceReader
	ReferenceWriter
}

// IDExtractor extracts object IDs from a given object
type IDExtractor func(Object) (ObjectID, error)

// NewDACGraph bootstraps a new graph using the given adapters
func NewDACGraph(objAd ObjectAdapter, refAd ReferenceAdapter) (*Graph, error) {
	return &Graph{ObjectAdapter: objAd, ReferenceAdapter: refAd}, nil
}

// FindLowestCommonAncestor traverses the graph recursively to find the lowest common ancestor of the given references
func (g *Graph) FindLowestCommonAncestor(refs ...string) (*Object, error) {
	if len(refs) < 2 {
		return nil, fmt.Errorf("Not enough references given to find ancestor: Found %v but need at least 2", len(refs))
	}

	// Extract the right reference and process errors or inexistent references
	var leftRef, found, err = g.ReferenceAdapter.ReadReference(refs[0])
	if err != nil {
		return nil, fmt.Errorf("Error while reading reference %s", refs[0])
	} else if !found {
		return nil, fmt.Errorf("Cannot find reference %s", refs[0])
	}

	var leftID = leftRef.TargetID
	var rightID ObjectID

	if len(refs) > 2 {
		var recLeft, err = g.FindLowestCommonAncestor(refs[1:]...)
		if err != nil {
			return nil, err
		} else if recLeft == nil {
			return nil, fmt.Errorf("Cannot find lowest common ancestor")
		} else {
			rightID = recLeft.ID
		}
	} else {
		var rightRef, found, err = g.ReferenceAdapter.ReadReference(refs[1])
		if err != nil {
			return nil, fmt.Errorf("Error while reading reference %s", refs[1])
		} else if !found {
			return nil, fmt.Errorf("Cannot find reference %s", refs[1])
		}
		rightID = rightRef.TargetID
	}

	var leftBacklog = make([]*tree.TreeNode, 1)
	var leftObjectsFound = map[ObjectID]int64{}
	leftBacklog[0] = &tree.TreeNode{ID: leftID, Depth: 0}

	for len(leftBacklog) > 0 {
		var currentItem = leftBacklog[0]
		var currentID = currentItem.ID.(ObjectID)

		if currentID == EmptyID {
			break
		}

		var currentObject, err = g.ObjectAdapter.ReadObject(currentID[:])
		if err != nil {
			return nil, err
		}

		for _, ancestor := range currentObject.PredecessorIDs {
			var ancestorNode = &tree.TreeNode{ID: ancestor, Depth: currentItem.Depth}
			leftBacklog = append([]*tree.TreeNode{ancestorNode}, leftBacklog...)
		}

		leftObjectsFound[currentObject.ID] = currentItem.Depth
	}

	var rightBacklog = make([]*tree.TreeNode, 1)
	var collisions = map[ObjectID]int64{}
	rightBacklog[0] = &tree.TreeNode{ID: rightID, Depth: 0}
	for len(rightBacklog) > 0 {
		var currentItem = rightBacklog[0]
		var currentID = currentItem.ID.(ObjectID)

		if currentID == EmptyID {
			break
		}

		var currentObject, err = g.ObjectAdapter.ReadObject(currentID[:])
		if err != nil {
			return nil, err
		}

		if _, exists := leftObjectsFound[currentObject.ID]; exists {
			collisions[currentObject.ID] = currentItem.Depth
		}

		for _, ancestor := range currentObject.PredecessorIDs {
			var ancestorNode = &tree.TreeNode{ID: ancestor, Depth: currentItem.Depth}
			rightBacklog = append([]*tree.TreeNode{ancestorNode}, leftBacklog...)
		}
	}

	if len(collisions) < 1 {
		return nil, fmt.Errorf("Cannot find a lowest common ancestor for the given objects %#x and %#x", leftID[:8], rightID[:8])
	}

	var shortestCollisionPoint ObjectID
	var shortestCollisionPathLenght int64
	shortestCollisionPathLenght = math.MaxInt64

	for k := range collisions {
		var totalPathLength = collisions[k] + leftObjectsFound[k]
		if totalPathLength < shortestCollisionPathLenght {
			shortestCollisionPathLenght = totalPathLength
			shortestCollisionPoint = k
		}
	}

	obj, err := g.ObjectAdapter.ReadObject(shortestCollisionPoint[:])
	if err != nil {
		return nil, fmt.Errorf("Cannot find object with id %#x", shortestCollisionPoint[:4])
	}

	return &obj, nil
}

// AppendNode adds a new node to the DAC given it's predecessor without moving any references
func (g *Graph) AppendNode(content []byte, predecessors ...ObjectID) (*Object, error) {
	var buf = make([]byte, 0)

	for _, predecessor := range predecessors {
		buf = append(buf, predecessor[:]...)
	}
	buf = append(buf, content...)

	h := ObjectID{}
	// Compute a 64-byte hash of buf and put it in h.
	sha3.ShakeSum128(h[:], buf)
	var obj = Object{}
	obj.ID = h
	obj.Content = content
	obj.PredecessorIDs = append([]ObjectID{}, predecessors...)

	g.ObjectAdapter.WriteObject(obj)

	return &obj, nil
}

// AppendNodeToRef appends a new node to the node specified in the given ref
func (g *Graph) AppendNodeToRef(content []byte, refName string) (*Object, error) {
	var ref, found, err = g.ReferenceAdapter.ReadReference(refName)
	if err != nil {
		return nil, err
	}

	newObj, err := g.AppendNode(content, []ObjectID{ref.TargetID}...)
	if err != nil {
		return nil, err
	}

	if !found {
		ref.TargetID = newObj.ID
	}

	g.ReferenceAdapter.WriteReference(ref)

	return newObj, nil
}

// Reference attaches a reference to the given object ID
func (g *Graph) Reference(objectID ObjectID, name string) (Reference, error) {
	var obj, err = g.ObjectAdapter.ReadObject(objectID[:])
	if err != nil {
		return Reference{}, err
	}

	var ref = Reference{
		Name:     name,
		TargetID: obj.ID}

	err = g.ReferenceAdapter.WriteReference(ref)
	if err != nil {
		return ref, err
	}

	return ref, nil
}
