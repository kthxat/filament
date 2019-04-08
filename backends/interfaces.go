package backends

import (
	"io"
	"os"
)

type Backend interface {
	// Close releases all used resources of this instance, including living
	// connections to the backend.
	Close() error
}

type Authenticator interface {
	Backend

	// Authenticate tests the given credentials against the backend, leaving
	// the connection open for later reuse.
	Authenticate(username, password string) (ok bool, err error)
}

type Storage interface {
	Backend

	// Stat returns statistics for the given file from the backend.
	Stat(path string) (os.FileInfo, error)

	// ReadDir returns a list of file information that are contained
	// in the given folder path.
	ReadDir(path string) ([]os.FileInfo, error)

	// Retrieve asks the backend for a file and writes its contents to the
	// destination writer.
	Retrieve(path string, dest io.Writer) error

	// IsLoggedInAs returns whether the same backend instance was already used
	// to authenticate as the given username. If this returns true, the
	// application will reuse this instance to access files.
	IsLoggedInAs(username string) bool
}
