# Filament

Access FTP-hosted files via HTTP. Filament reads a directory tree over FTP and
serves it as plain HTML listings. Visitors authenticate with HTTP Basic auth,
and Filament checks those credentials against the FTP server.

Status: one backend (FTP), no TLS on the HTTP side yet, and a minimal
server-rendered frontend.

## Quick start

Write a `filament.toml`:

```toml
[HTTP]
    ListenAddress = ":8080"
    AuthenticationRealm = "Filament"

[Backends.ftp]
    URL = "ftpes://ftp.example.com:21/"
    Timeout = "30s"
```

Name the backend section after the backend's ID. Filament looks the
configuration up by ID, so a section called `[Backends.ftp1]` is ignored and
the backend starts without a URL.

Run it:

```bash
docker run --rm -p 8080:8080 -v "$PWD:/config" ghcr.io/kthxat/filament:main
```

The image reads `/config/filament.toml` and runs as an unprivileged user. Open
<http://localhost:8080>; the browser shows a Basic auth prompt, and your FTP
credentials get you in.

## Configuration

Config files are searched for in this order, first match wins:

1. `$XDG_CONFIG_HOME/filament/filament.toml`
2. `~/.filament/filament.toml`
3. `/etc/filament/filament.toml` (not on Windows)
4. `./filament.toml`

The container image sets its working directory to `/config`, which is why the
example above mounts there.

Any setting can also come from the environment, prefixed with `FILAMENT_` and
uppercased, with `_` marking nesting:

```bash
FILAMENT_HTTP_LISTENADDRESS=":9090"
```

An environment variable only takes effect for a key that already has a built-in
default or appears in the config file, because Filament unmarshals the whole
config at once. `HTTP.ListenAddress` has a default, so it works on its own; a
key that appears nowhere in the file does not.

### HTTP

| Setting | Description |
| --- | --- |
| `ListenAddress` | Address to listen on, defaults to `:8080`. |
| `AuthenticationRealm` | Realm reported in the `WWW-Authenticate` header. |

TLS for the HTTP listener is not implemented. Put Filament behind a reverse
proxy if you need HTTPS.

### FTP backend

| Setting | Description |
| --- | --- |
| `URL` | FTP endpoint. The scheme picks the TLS mode: `ftp://` plain, `ftps://` implicit TLS, `ftpes://` explicit TLS (AUTH TLS). |
| `Timeout` | Connection timeout, e.g. `"30s"`. |
| `TLSServerName` | Server name for TLS verification, defaults to the host from `URL`. |
| `InsecureSkipVerify` | Skip TLS certificate verification. |
| `ActiveTransfers` | Use active (PORT) transfers instead of passive. |
| `ActiveListenAddr` | Local address to bind for active transfers. |
| `DisableEPSV` | Disable extended passive mode. |
| `IPv6Lookup` | Prefer AAAA lookups. |
| `ServerLocation` | IANA timezone name to interpret FTP timestamps with. |

Credentials in the `URL` do not log anyone in. Each request opens its own FTP
connection as the visitor, using the credentials from the auth prompt.

## Usage

- Directories sort first, then files, both alphabetically, with sizes
  humanized.
- Request `<directory>/.filament/archive.zip` to download a whole tree as a ZIP
  archive. The same path with a `.tar`, `.tar.gz`, `.tar.bz2`, `.tar.xz` or
  `.7z` suffix is recognized but returns `501 Not Implemented`.
- Sessions live in memory and expire after five minutes of inactivity. The same
  credentials reuse an existing session, and Filament keeps the password only
  as a bcrypt hash.

## Building from source

Requires Go 1.23 or newer.

```bash
go generate ./...
go build -v .
./filament
```

Run `go generate` first. It writes `rice-box.go`, which embeds the HTML
templates and is not checked in, so a build without it has no templates.

## Container images

The Docker workflow builds `linux/amd64`, `linux/arm64`, `linux/arm/v7` and
`linux/arm/v6`, and pushes `ghcr.io/kthxat/filament` tagged `:<branch>` and
`:sha-<commit>` on every push, plus semver tags for `v*` git tags.

## License

MIT, see [LICENSE.txt](LICENSE.txt).
