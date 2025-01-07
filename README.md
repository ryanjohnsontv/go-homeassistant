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
    "log"
    "os"
    "time"
    
    "github.com/joho/godotenv"
    "github.com/ryanjohnsontv/go-homeassistant/rest"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file", err)
    }

    client, err := rest.NewClient(os.Getenv("HA_HOST"), os.Getenv("HA_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    state, err := client.GetState("light.living_room")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Light state:", response)

    if _, err := client.CallService(domains.Light, "toggle", nil); err != nil {
        log.Fatal(err)
    }
}
```

### WebSocket Client

The WebSocket client supports real-time communication with Home Assistant.

#### Websocket Example

```go
package main

import (
    "log"
    "os"
    "time"

    "github.com/joho/godotenv"
    "github.com/ryanjohnsontv/go-homeassistant/shared/actions/light"
    "github.com/ryanjohnsontv/go-homeassistant/websocket"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file", err)
    }

    client, err := websocket.NewClient(os.Getenv("HA_HOST"), os.Getenv("HA_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    if err := client.Run(); err != nil {
        return err
    }

    client.Actions.Light.TurnOn().Entities("light.guest_bedroom_floor_lamp").ServiceData(light.TurnOnData{
        Transition:    5,
        ColorName:     "red",
        BrightnessPct: 100,
    }).Execute()

    time.Sleep(15 * time.Second)

    client.Actions.Light.TurnOff().Entities("light.guest_bedroom_floor_lamp").Execute()

    time.Sleep(15 * time.Second)
}
```

## Development

### Prerequisites

- Go 1.23+

### Linting

Lint the project using the configuration in `.golangci.yml`:

```bash
make lint
```

## Contributing

Contributions are welcome! To get started:

1. Fork the repository.
2. Create a feature branch.
3. Commit your changes and open a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
