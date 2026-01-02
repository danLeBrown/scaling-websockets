# Load-Balanced WebSockets

A hobby project demonstrating horizontal scaling of WebSocket connections across multiple application instances using HAProxy load balancing and Redis pub/sub for message distribution.

## Overview

This project showcases how to scale WebSocket applications horizontally by:
- Distributing WebSocket connections across multiple Go application instances
- Using HAProxy as a load balancer to route connections
- Leveraging Redis pub/sub to synchronize messages across all instances
- Ensuring users connected to different instances can still communicate in real-time

## Architecture

```
                    ┌─────────────┐
                    │   HAProxy   │
                    │  (Port 8080)│
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐       ┌────▼────┐       ┌────▼────┐
   │ lbchat1 │       │ lbchat2 │       │ lbchat3 │
   │ (App 1) │       │ (App 2) │       │ (App 3) │
   └────┬────┘       └────┬────┘       └────┬────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                    ┌──────▼──────┐
                    │    Redis     │
                    │  (Pub/Sub)   │
                    └──────────────┘
```

### Components

- **HAProxy**: Load balancer distributing HTTP and WebSocket connections across backend servers
- **Go Application Instances** (lbchat1-4): Multiple instances of the WebSocket server, each with a unique APPID
- **Redis**: Message broker using pub/sub to synchronize messages across all instances
- **Frontend**: Modern chat interface built with Vue.js and Tailwind CSS

## Features

- ✅ **Horizontal Scaling**: Multiple application instances handle WebSocket connections
- ✅ **Cross-Instance Messaging**: Messages sent to one instance are broadcast to all instances via Redis
- ✅ **Load Balancing**: HAProxy distributes connections using round-robin
- ✅ **Instance Identification**: Each instance displays its APPID in the UI
- ✅ **Real-Time Chat**: WebSocket-based chat with username management
- ✅ **Modern UI**: Responsive design with Tailwind CSS
- ✅ **Health Checks**: `/health` endpoint for monitoring

## Tech Stack

- **Backend**: Go 1.23.1
- **WebSocket**: Gorilla WebSocket
- **Message Broker**: Redis 6.2.19 (pub/sub)
- **Load Balancer**: HAProxy 3.3
- **Frontend**: Vue.js 3, Tailwind CSS
- **Containerization**: Docker & Docker Compose

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)

### Running the Application

1. **Clone the repository** (if applicable)

2. **Start all services**:
   ```bash
   docker compose up --build
   ```

3. **Access the application**:
   - Open your browser to `http://localhost:8080/chat-app`
   - The load balancer will route you to one of the four application instances
   - Check the APPID badge in the header to see which instance is serving you

4. **Test scaling**:
   - Open multiple browser tabs/windows to `http://localhost:8080/chat-app`
   - Notice how different instances serve different connections
   - Send messages from different tabs - they should appear in all tabs regardless of which instance they're connected to

### Development

To run locally without Docker:

```bash
# Set environment variables
export APPID=1111
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0

# Run the application
go run main.go
```

## How It Works

### WebSocket Connection Flow

1. Client connects to `ws://localhost:8080/ws`
2. HAProxy routes the connection to one of the backend instances (lbchat1-4)
3. The Go application upgrades the HTTP connection to WebSocket
4. The connection is stored in the instance's local client map

### Message Broadcasting Flow

1. Client sends a message via WebSocket
2. The receiving instance publishes the message to Redis pub/sub channel `chat-app`
3. All instances (including the sender) subscribe to the Redis channel
4. Each instance receives the message and broadcasts it to its local WebSocket clients
5. All connected clients receive the message, regardless of which instance they're connected to

### Load Balancing

HAProxy uses round-robin to distribute connections:
- First connection → lbchat1
- Second connection → lbchat2
- Third connection → lbchat3
- Fourth connection → lbchat4
- Fifth connection → lbchat1 (cycles back)

## Project Structure

```
.
├── compose.yaml          # Docker Compose configuration
├── Dockerfile            # Multi-stage build for Go application
├── main.go               # Go WebSocket server with Redis pub/sub
├── go.mod                # Go dependencies
├── haproxy/
│   └── haproxy.cfg       # HAProxy load balancer configuration
└── public/
    └── index.html        # Frontend chat interface
```

## Configuration

### Environment Variables

Each application instance can be configured via environment variables:

- `APPID`: Unique identifier for the instance (displayed in UI)
- `REDIS_HOST`: Redis server hostname
- `REDIS_PORT`: Redis server port (default: 6379)
- `REDIS_PASSWORD`: Redis password (empty by default)
- `REDIS_DB`: Redis database number (default: 0)

### Scaling

To add more instances:

1. Add a new service in `compose.yaml`:
   ```yaml
   lbchat5:
     <<: *app-build
     depends_on:
       - redis
     environment:
       - APPID=5555
       - REDIS_HOST=redis
       - REDIS_PORT=6379
       - REDIS_PASSWORD=
       - REDIS_DB=0
   ```

2. Add the server to `haproxy/haproxy.cfg`:
   ```
   server s5 lbchat5:3000
   ```

3. Restart the services:
   ```bash
   docker compose up -d
   ```

## API Endpoints

- `GET /chat-app` - Serves the chat interface HTML
- `GET /ws` - WebSocket endpoint for real-time messaging
- `GET /health` - Health check endpoint (returns "OK")

## Message Format

Messages are JSON-encoded:

```json
{
  "username": "string",
  "message": "string",
  "timestamp": "ISO8601 string"
}
```

## Notes

- This is a demonstration project for learning WebSocket scaling patterns
- The current implementation uses Vue.js for the frontend (originally planned with HTMX)
- Redis pub/sub ensures message consistency across instances
- HAProxy handles WebSocket upgrade and connection persistence
- Each instance maintains its own local client map for efficient message delivery

## License

This is a hobby/educational project.

