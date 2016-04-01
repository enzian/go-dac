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
		var recLeft, recErr = g.FindLowestCommonAncestor(refs[1:]...)
		if recErr != nil {
			return nil, recErr
		} else if recLeft == nil {
			return nil, fmt.Errorf("Cannot find lowest common ancestor")
		} else {
			rightID = recLeft.ID
		}
	} else {
		var rightRef, found, refErr = g.ReferenceAdapter.ReadReference(refs[1])
		if refErr != nil {
			return nil, fmt.Errorf("Error while reading reference %s", refs[1])
		} else if !found {
			return nil, fmt.Errorf("Cannot find reference %s", refs[1])
		}
		rightID = rightRef.TargetID
	}

	// Parses the graph to a tree beginning at the specified id
	leftNodes, _, err := g.toTree(leftID)

	// Find all intersection object with leftNodes
	// Function that analyzes whether the given object represents a collision
	var isCollision = func(obj *Object) bool {
		var _, exists = leftNodes[obj.ID]
		return exists
	}

	// Records the node where a collision happens
	var collisions = []*tree.TreeNode{}
	var collisionRecorder = func(node *tree.TreeNode) {
		collisions = append(collisions, node)
	}

	_, _, err = g.toCollisionTerminatedTree(rightID, isCollision, collisionRecorder)

	var shortestCollisionPoint ObjectID
	var shortestCollisionPathLenght int64
	shortestCollisionPathLenght = math.MaxInt64

	// Iterate over all collisions and find the one with the shortest path length
	for _, k := range collisions {
		var id = ObjectID(k.ID.(ObjectID))
		var totalPathLength = k.Depth + leftNodes[id].Depth
		if totalPathLength < shortestCollisionPathLenght {
			shortestCollisionPathLenght = totalPathLength
			shortestCollisionPoint = id
		}
	}

	obj, err := g.ObjectAdapter.ReadObject(shortestCollisionPoint[:])
	if err != nil {
		return nil, fmt.Errorf("Cannot find object with id %#x", shortestCollisionPoint[:4])
	}

	return &obj, nil
}

func (g *Graph) toTree(startPoint ObjectID) (map[ObjectID]*tree.TreeNode, *tree.TreeNode, error) {
	var isCollision = func(obj *Object) bool { return false }
	var collisionRecorder = func(node *tree.TreeNode) {}

	var foundItems, rootNode, err = g.toCollisionTerminatedTree(startPoint, isCollision, collisionRecorder)

	return foundItems, rootNode, err
}

func (g *Graph) toCollisionTerminatedTree(startPoint ObjectID, detectCollision func(obj *Object) bool, collisionRecorder func(treeNode *tree.TreeNode)) (map[ObjectID]*tree.TreeNode, *tree.TreeNode, error) {
	// The backlog holds all nodes that still need to be processed!
	var backlog = make([]*tree.TreeNode, 1)
	// The first item in the backlog is the root node
	var rootNode = &tree.TreeNode{ID: startPoint, Depth: 0}
	backlog[0] = rootNode

	var foundItems = make(map[ObjectID]*tree.TreeNode)

	// Iterate on the backlong as long as there are items in it!
	for len(backlog) > 0 {
		// The current item is always the first item in the backlog
		var currentItem = backlog[0]
		// Extract the ObjectID of the object to be processed from the treeNodes Id field
		var currentID = currentItem.ID.(ObjectID)

		// Stop iterating if we are at the end of the graph
		if currentID == EmptyID {
			break
		}

		// read the actual object for the given object id
		var currentObject, readErr = g.ObjectAdapter.ReadObject(currentID[:])
		if readErr != nil {
			return nil, nil, readErr
		}

		if detectCollision(&currentObject) {
			collisionRecorder(currentItem)
		}

		// all ancestors of an object will be added to the backlog for further processing
		for _, ancestor := range currentObject.PredecessorIDs {
			var ancestorNode = currentItem.AppendChild(ancestor)
			backlog = append([]*tree.TreeNode{ancestorNode}, backlog...)
		}

		foundItems[currentID] = currentItem
	}

	return foundItems, rootNode, nil
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
