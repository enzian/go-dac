package dac

import (
	"fmt"

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
	Read(id []byte) (Object, error)
}

// ObjectWriter persist new/changed objects
type ObjectWriter interface {
	Write(obj Object) error
}

// ObjectAdapter extends read and write capabilities for objects to the graph
type ObjectAdapter interface {
	ObjectReader
	ObjectWriter
}

// ReferenceReader reads references
type ReferenceReader interface {
	Read(name string) (Reference, error)
}

// ReferenceWriter persist new/changed references
type ReferenceWriter interface {
	Write(obj Reference) error
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

	if len(refs) > 2 {
		g.FindLowestCommonAncestor(refs[1:]...)
	}

	return nil, nil
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

	g.ObjectAdapter.Write(obj)

	return &obj, nil
}

// AppendNodeToRef appends a new node to the node specified in the given ref
func (g *Graph) AppendNodeToRef(content []byte, ref string) (*Object, error) {
	var foundRef, err = g.ReferenceAdapter.Read(ref)
	if err != nil {
		return nil, err
	}

	newObj, err := g.AppendNode(content, []ObjectID{foundRef.TargetID}...)
	if err != nil {
		return nil, err
	}

	return newObj, nil
}
