package ftp

import (
	"github.com/kthxat/filament/backends"
	"github.com/secsy/goftp"
	"go.uber.org/multierr"
)

func (b *FTPBackend) Authenticate(username, password string) (ok bool, err error) {
	if b.client != nil {
		err = b.client.Close()
		b.client = nil
	}

	config := b.configTemplate
	config.User = username
	config.Password = password

	var dialErr error
	b.client, dialErr = goftp.DialConfig(config, b.configuredHost)
	err = multierr.Append(err, dialErr)

	b.authenticatedUsername = username
	ok = true
	return
}

func (b *FTPBackend) ChangePassword(newPassword string) (err error) {
	err = backends.ErrUnsupportedOperation
	return
}
