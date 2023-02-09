package backends

import (
	"fmt"
	"sync"
)

var (
	backends     = []*BackendDescriptor{}
	backendsLock sync.Mutex
)

// Register registers a backend using the given descriptor.
func Register(descriptor *BackendDescriptor) {
	backendsLock.Lock()
	defer backendsLock.Unlock()

	if unsyncedGetByID(descriptor.ID) != nil {
		panic(fmt.Errorf("tried to register backend %s more than once", descriptor.ID))
	}
	backends = append(backends, descriptor)
}

func unsyncedGetByID(id string) (backendDescriptor *BackendDescriptor) {
	for _, foundDescriptor := range backends {
		if foundDescriptor.ID == id {
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
