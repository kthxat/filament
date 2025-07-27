# Filament

Filament is a lightweight HTTP-to-FTP server that provides a web interface for accessing files on FTP servers. It acts as a bridge, allowing you to browse and download files from FTP servers through a simple web interface.

## Features

- **Web-based FTP access**: Browse FTP servers through your web browser
- **Multi-architecture support**: Runs on amd64, arm64, armv7, and armv6
- **Docker ready**: Pre-built Docker images available
- **Configurable**: Flexible configuration system with TOML support
- **Authentication**: Built-in authentication and session management
- **Lightweight**: Minimal resource footprint

## Quick Start

### Using Docker (Recommended)

The easiest way to run Filament is using Docker:

```bash
docker run -p 8080:8080 ghcr.io/kthxat/filament:latest
```

This will start Filament on port 8080 with default settings.

### Using Pre-built Releases

Download the latest release from [GitHub Packages](https://github.com/kthxat/filament/pkgs/container/filament):

```bash
# Pull the latest image
docker pull ghcr.io/kthxat/filament:latest
```

## Configuration

Filament uses a configuration file to set up FTP backends and HTTP server settings. Create a configuration file named `filament.toml`:

```toml
# HTTP server configuration
[HTTP]
ListenAddress = ":8080"
AuthenticationRealm = "Filament FTP Access"

# FTP backend configuration
[Backends.ftp1]
URL = "ftp://username:password@ftp.example.com:21/"
Timeout = "30s"
IPv6Lookup = false
ActiveTransfers = false
InsecureSkipVerify = false
```

### Configuration File Locations

Filament will look for the configuration file in the following locations (in order):

1. `$XDG_CONFIG_HOME/filament/filament.toml`
2. `~/.filament/filament.toml`
3. `/etc/filament/filament.toml` (Unix/Linux only)
4. `./filament.toml` (current directory)

### Environment Variables

You can also configure Filament using environment variables with the `FILAMENT_` prefix:

```bash
export FILAMENT_HTTP_LISTENADDRESS=":9090"
export FILAMENT_AUTHENTICATIONBACKEND="ftp1"
```

## Installation

### Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'
services:
  filament:
    image: ghcr.io/kthxat/filament:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config:/config
    environment:
      - FILAMENT_HTTP_LISTENADDRESS=:8080
```

Then run:

```bash
docker-compose up -d
```

### Building from Source

Requirements:
- Go 1.23.0 or later

```bash
# Clone the repository
git clone https://github.com/kthxat/filament.git
cd filament

# Build
go build -v .

# Run
./filament
```

## Usage

1. Start Filament with your configuration
2. Open your web browser and navigate to `http://localhost:8080`
3. Log in using your FTP credentials (configured in the backend)
4. Browse and download files through the web interface

## Releases

- **Docker Images**: Available on [GitHub Container Registry](https://github.com/kthxat/filament/pkgs/container/filament)
- **Source Releases**: Check the [releases page](https://github.com/kthxat/filament/releases) for source archives

## Development

### Prerequisites

- Go 1.23.0+
- Docker (for containerized development)

### Building

```bash
# Build the application
go build -v .

# Run with default configuration
./filament
```

### Docker Development

```bash
# Build Docker image
docker build -t filament:dev .

# Run development container
docker run -p 8080:8080 filament:dev
```
