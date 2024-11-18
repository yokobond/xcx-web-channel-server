# WebSocket Channel Server

This is a WebSocket server for [Web Channel Extension for Xcratch](https://yokobond.github.io/xcx-web-channel/).

A lightweight WebSocket server implementing a publish/subscribe (pub/sub) pattern for real-time message broadcasting. The server supports both secure (WSS) and non-secure (WS) WebSocket connections.

## Features

- WebSocket-based pub/sub messaging system
- Support for both WS and WSS protocols
- Topic-based message routing
- Configurable allowed origins to restrict connections
- Auto-reconnection for clients
- Web-based test client interface
- Periodic connection health checks (ping/pong)
- Message repeat functionality for testing

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yokobond/xcx-web-channel-server.git
   ```

2. Build the server:

   ```bash
   go build -o web-channel-server ./cmd/web-channel-server
   ```

## Usage

### Starting the Server

The server can run without a `config.json` file. If the configuration file is missing, default settings are used.

Start the server with WSS enabled:

```bash
./web-channel-server -wss -config config.json
```

## Configuration

Create a `config.json` file for server settings:

```json
{
    "wsPort": 8080,
    "wssPort": 8443,
    "certFile": "/path/to/cert.pem",
    "keyFile": "/path/to/key.pem",
    "allowedOrigins": ["http://example.com", "http://anotherdomain.com"]
}
```

If `allowedOrigins` is empty or not specified, the server allows connections from any origin.

## Web Client

A web-based test client is available at `web/web-channel-server/index.html`. Open this file in a browser to:

- Subscribe to topics
- Publish messages
- Test repeated message sending
- Monitor connection status

## WebSocket API
Messages use JSON format:

Subscribe to a topic:

```json
{
    "action": "subscribe",
    "topic": "example-topic"
}
```

Publish a message:

```json
{
    "action": "publish",
    "topic": "example-topic",
    "message": "Hello, World!"
}
```

## Development
The project structure:

```
xcx-web-channel-server/
├── cmd/
│   └── web-channel-server/
│       ├── main.go
│       ├── client.go
│       └── config.go
├── web/
│   └── web-channel-server/
│       └── index.html
├── config.json     // optional configuration file
└── README.md
```

## License
MIT License
