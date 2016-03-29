package dac

// Object represents a node in the directed acyclic graph
type Object struct {
	ID            []ObjectID
	Content       []byte
	PredecessorID []ObjectID
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
	Read(id []byte) (Reference, error)
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
func NewDACGraph(objAd ObjectAdapter, refAd ReferenceAdapter) (Graph, error) {
	return Graph{ObjectAdapter: objAd, ReferenceAdapter: refAd}, nil
}
