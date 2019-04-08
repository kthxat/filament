package backend_ftp

import (
	"github.com/secsy/goftp"

	"git.kthx.at/icedream/filament/backends"
)

func (b *FTPBackend) Authenticate(username, password string) (ok bool, err error) {
	if b.client != nil {
		b.client.Close()
		b.client = nil
	}

	config := b.configTemplate
	config.User = username
	config.Password = password

	b.client, err = goftp.DialConfig(config, b.configuredHost)
	if err != nil {
		return
	}

	b.authenticatedUsername = username
	ok = true
	return
}

func (b *FTPBackend) ChangePassword(newPassword string) (err error) {
	err = backends.ErrUnsupportedOperation
	return
}
