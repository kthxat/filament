package backend_ftp

import (
	"io"
	"os"
)

func (b *FTPBackend) IsLoggedInAs(username string) bool {
	return username == b.authenticatedUsername
}

func (b *FTPBackend) Stat(path string) (info os.FileInfo, err error) {
	info, err = b.client.Stat(path)
	return
}

func (b *FTPBackend) ReadDir(path string) (info []os.FileInfo, err error) {
	info, err = b.client.ReadDir(path)
	return
}

func (b *FTPBackend) Retrieve(path string, w io.Writer) (err error) {
	err = b.client.Retrieve(path, w)
	return
}
