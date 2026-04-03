# Peer-CDN

A peer-assisted Content Delivery Network (CDN) built in Go using the BitTorrent protocol for efficient, distributed file delivery.

## Overview

Peer-CDN is a decentralized content delivery system that leverages peer-to-peer technology to distribute files efficiently. By combining BitTorrent protocol with a Redis-based tracker, it enables concurrent piece distribution with automatic fallback to origin servers, making it ideal for high-scale content delivery scenarios.

## Features

- 🚀 **BitTorrent Protocol**: Efficient peer-to-peer file distribution
- 📍 **Redis-Based Tracker**: Centralized peer management and discovery
- ⚙️ **Concurrent Piece Distribution**: Parallel downloads from multiple peers
- 🔄 **Origin Fallback**: Automatic fallback to origin servers when peers are unavailable
- 🔧 **Built in Go**: High performance, concurrent, and efficient implementation
- 🌐 **Scalable**: Designed to handle large-scale content delivery

## Architecture

The project is organized into the following main components:

```
peer-cdn/
├── cmd/              # Command-line utilities and entry points
├── api/              # API definitions and handlers
├── internal/         # Core implementation packages
├── go.mod            # Go module definition
└── go.sum            # Dependency checksums
```

### Key Directories

- **`cmd/`**: Contains executable applications and CLI tools for running the CDN components
- **`api/`**: Defines API interfaces and request/response handlers
- **`internal/`**: Core business logic including BitTorrent protocol implementation, tracker management, and piece distribution

## Getting Started

### Prerequisites

- Go 1.19 or higher
- Redis (for tracker functionality)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/NeerajRijhwani/peer-cdn.git
cd peer-cdn
```

2. Install dependencies:

```bash
go mod download
```

3. Build the project:

```bash
go build ./cmd/...
```

## Configuration

The system uses Redis for tracking peer information and managing piece distribution metadata. Configure your Redis connection before running the CDN components.

## Usage

### Running the CDN

```bash
# Example command to run the CDN server
./peer-cdn [options]
```

For available options and detailed usage instructions, refer to the command-line help:

```bash
./peer-cdn --help
```

## Project Status

This is an actively maintained project. The repository was created 50 days ago and is continuously being developed.

## Technologies Used

- **Language**: Go
- **Protocol**: BitTorrent
- **Cache/Storage**: Redis
- **Architecture**: Distributed, Peer-to-Peer

## Contributing

Contributions are welcome! Please feel free to:

- Report bugs and issues
- Suggest features and improvements
- Submit pull requests

## License

This project is open source. Check the LICENSE file for more details.

## Contact & Support

For questions, issues, or discussions, please open an issue on the GitHub repository.

---

**Repository**: [NeerajRijhwani/peer-cdn](https://github.com/NeerajRijhwani/peer-cdn)

Last Updated: April 3, 2026
