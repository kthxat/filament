package backends

import (
	"fmt"
	"sync"
)

var backends = []*BackendDescriptor{}
var backendsLock sync.Mutex

// Register registers a backend using the given descriptor.
func Register(descriptor *BackendDescriptor) {
	backendsLock.Lock()
	defer backendsLock.Unlock()

	if unsyncedGetByID(descriptor.Id) != nil {
		panic(fmt.Errorf("Tried to register backend %s more than once.", descriptor.Id))
	}
	backends = append(backends, descriptor)
}

func unsyncedGetByID(id string) (backendDescriptor *BackendDescriptor) {
	for _, foundDescriptor := range backends {
		if foundDescriptor.Id == id {
			backendDescriptor = foundDescriptor
			return
		}
	}
	return
}

// GetByID returns the backend descriptor which matches the wanted ID.
// If no such backend exists, nil is returned.
func GetByID(id string) (backendDescriptor *BackendDescriptor) {
	backendsLock.Lock()
	defer backendsLock.Unlock()

	backendDescriptor = unsyncedGetByID(id)
	return
}

// GetAll returns an array of all backend descriptors.
func GetAll() (backendDescriptors []*BackendDescriptor) {
	backendsLock.Lock()
	defer backendsLock.Unlock()

	backendDescriptors = make([]*BackendDescriptor, len(backends))
	copy(backendDescriptors, backends)
	return
}
