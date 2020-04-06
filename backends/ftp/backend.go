package backend_ftp

import (
	"crypto/tls"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/kthxat/filament/backends"
	"github.com/secsy/goftp"
)

type FTPBackendConfiguration struct {
	URL     string
	Timeout time.Duration

	// TODO - TLSConfig        *tls.Config
	IPv6Lookup         bool
	ActiveTransfers    bool
	ActiveListenAddr   string
	DisableEPSV        bool
	ServerLocation     string
	InsecureSkipVerify bool
	TLSServerName      string
}

func (c *FTPBackendConfiguration) makeFTPClientConfig() (retval goftp.Config, err error) {
	retval = goftp.Config{
		IPv6Lookup:       c.IPv6Lookup,
		ActiveTransfers:  c.ActiveTransfers,
		ActiveListenAddr: c.ActiveListenAddr,
		DisableEPSV:      c.DisableEPSV,
	}
	if len(c.ServerLocation) > 0 {
		retval.ServerLocation, err = time.LoadLocation(c.ServerLocation)
		if err != nil {
			return
		}
	}
	return
}

func init() {
	backends.Register(&backends.BackendDescriptor{
		Id:          "ftp",
		DisplayName: "FTP",
		Type:        reflect.TypeOf(new(FTPBackend)),
		New: func(params *backends.BackendConstructionParams) (backend backends.Backend, err error) {
			config := new(FTPBackendConfiguration)
			err = params.Config.Unmarshal(config)
			if err != nil {
				return
			}

			ftpConfig, err := config.makeFTPClientConfig()
			if err != nil {
				return
			}

			// determine which TLS mode to use from URL
			ftpUrl, err := url.Parse(config.URL)
			if err != nil {
				return
			}
			switch strings.ToLower(ftpUrl.Scheme) {
			case "ftps":
				// implicit
				ftpConfig.TLSMode = goftp.TLSImplicit
				ftpConfig.TLSConfig = new(tls.Config)
			case "ftpes":
				// explicit
				ftpConfig.TLSMode = goftp.TLSExplicit
				ftpConfig.TLSConfig = new(tls.Config)
			}
			if ftpConfig.TLSConfig != nil {
				ftpConfig.TLSConfig.InsecureSkipVerify = config.InsecureSkipVerify
				if !config.InsecureSkipVerify {
					if len(config.TLSServerName) > 0 {
						ftpConfig.TLSConfig.ServerName = config.TLSServerName
					} else {
						ftpConfig.TLSConfig.ServerName = ftpUrl.Host
					}
				}
			}
			if ftpUrl.User != nil {
				ftpConfig.User = ftpUrl.User.Username()
				if pw, ok := ftpUrl.User.Password(); ok {
					ftpConfig.Password = pw
				}
			}

			backend = &FTPBackend{
				configTemplate: ftpConfig,
				configuredHost: ftpUrl.Host,
			}

			return
		},
	})
}

type FTPBackend struct {
	authenticatedUsername string
	client                *goftp.Client
	configTemplate        goftp.Config
	configuredHost        string
}

func (b *FTPBackend) Close() (err error) {
	if b.client != nil {
		err = b.client.Close()
		b.client = nil
	}
	return
}
