package transport

import "github.com/go-msf/go-dac"

// CommonObjectNegotiator is used to negotiate the objects that the client needs
type CommonObjectNegotiator func(g dac.Graph) ([]dac.ObjectID, error)

type ReferenceFetcher func() (map[string]dac.ObjectID, error)

// PackSender is used to send all objects the clients needs
type PackSender func(ids []dac.ObjectID) error

type DacClient func() (CommonObjectNegotiator, ReferenceFetcher)

// RefAdvertiser returns the receiving sides references
type RefAdvertiser func(graph dac.Graph) ([]dac.Reference, error)

type PackReceiver func() ([]dac.Object, error)

type ObjectChecker func(id dac.ObjectID) (bool, error)

type DacRemote func() (RefAdvertiser, ObjectChecker, PackReceiver)

// Used to send Objects for specific references
func SendObjects(graph dac.Graph, client DacClient) {
	// use the ReferenceFetcher to fetch all objects on the server side

	// Check that the reference is actually ahead of what objects the server advertiesed

	// backtrack all references to the objects advertised by the server

	// Submit these objects using the PackSender func
}

// Used to receive objects
func ReceiveObjects(graph dac.Graph, remote DacRemote) {
	// Advertise my references

	// ACK/NACK the objects IDS advertised by the sender

	// Use the PackReceiver to read all objects that need to be added to my repo
}
