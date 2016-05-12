package memory

import (
	"fmt"

	"github.com/go-msf/go-dac"
)

type MemoryAdapter struct {
	objects map[string]dac.Object
	refs    map[string]dac.Reference
}

func NewMemoryAdapter() dac.GraphAdapter {
	var adapter = new(MemoryAdapter)
	adapter.objects = map[string]dac.Object{}
	adapter.refs = map[string]dac.Reference{}

	return adapter
}

func (wr MemoryAdapter) WriteObject(obj dac.Object) error {
	wr.objects[string(obj.ID[:])] = obj
	return nil
}

func (wr MemoryAdapter) ReadObject(id []byte) (dac.Object, error) {
	var obj, found = wr.objects[string(id)]
	if !found {
		return dac.Object{}, fmt.Errorf("Cannot find %#x", id)
	}

	return obj, nil
}

func (wr MemoryAdapter) ReadReference(id string) (dac.Reference, bool, error) {
	var ref, found = wr.refs[id]
	if !found {
		return dac.Reference{Name: id}, found, nil
	}

	return ref, true, nil
}

func (wr MemoryAdapter) WriteReference(ref dac.Reference) error {
	wr.refs[ref.Name] = ref
	return nil
}
