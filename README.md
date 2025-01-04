# go-homeassistant

Clients for communicating with Home Assistant's REST and WebSocket APIs.

> :warning: **Under rapid and unplanned development**: This is a slow-maturing passion project!

## Features

- **REST API Integration**: Easily interact with Home Assistant's REST endpoints.
- **WebSocket API Support**: Establish persistent WebSocket connections for real-time updates.
- **Modular Design**: Reusable and well-organized components in `shared` for common functionality.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [REST Client](#rest-client)
  - [WebSocket Client](#websocket-client)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Installation

```bash
# Clone the repository
git clone https://github.com/ryanjohnsontv/go-homeassistant.git
cd go-homeassistant

# Install dependencies
go mod tidy
```

## Usage

### REST Client

The REST client allows you to interact with Home Assistant's REST API.

#### REST Example

```go
package main

import (
    "fmt"
    "github.com/ryanjohnsontv/go-homeassistant/rest"
)

func main() {
    client, err := rest.NewClient("http://homeassistant.local:8123", "your-access-token")
    if err != nil {
        return
    }
    response, err := client.GetState("light.living_room")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("Light state:", response)
}
```

### WebSocket Client

The WebSocket client supports real-time communication with Home Assistant.

#### Websocket Example

```go
package main

import (
    "fmt"
    "github.com/ryanjohnsontv/go-homeassistant/websocket"
)

func main() {
    client := websocket.NewClient("ws://homeassistant.local:8123/api/websocket", "your-access-token")
    if err := client.Connect(); err != nil {
        fmt.Println("Error connecting to WebSocket:", err)
        return
    }

    for message := range client.Listen() {
        fmt.Println("Received message:", message)
    }
}
```

## Development

### Prerequisites

- Go 1.23+

### Build and Run

```bash
# Build the project
make build

# Run tests
make test
```

### Linting

Lint the project using the configuration in `.golangci.yml`:

```bash
golangci-lint run
```

## Contributing

Contributions are welcome! To get started:

1. Fork the repository.
2. Create a feature branch.
3. Commit your changes and open a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
