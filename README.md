# Pynezzentials

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pynezz/pynezzentials)
[![Go Report Card](https://goreportcard.com/badge/github.com/pynezz/pynezzentials)](https://goreportcard.com/report/github.com/pynezz/pynezzentials)
[![License](https://img.shields.io/github/license/pynezz/pynezzentials)](LICENSE)

Pynezzentials is a Go package that provides some utilities for my personal projects. It is not intended to be a general-purpose library, but it may be useful for some people.

## Features

- **cryptoutils**: A set of utilities for encrypting and decrypting data.
- **fsUtils**: A set of utilities for working with files.
- **ansi**: A set of utilities for working with ANSI escape codes.

## Installation

To install Pynezzentials, run:

```bash
go get -u github.com/pynezz/pynezzentials
```

## Usage

```go
package main

import (
    "fmt"
    utils "github.com/pynezz/pynezzentials"
)

func main() {
    fmt.Println(utils.Bold("Hello, world!"))
}
```

## IPC
UNIX Domain sockets inter process communication (IPC) is a mechanism for exchanging data between processes running on the same host operating system. Pynezzentials provides a simple way to create a server and client for IPC.

### Server

```go
package main

import (
    "fmt"
    "github.com/pynezz/pynezzentials/ipc"
)

func main() {
    server := ipc.NewIPCServer("servername", [4]byte{0, 0, 0, 1})  // Identifier for the server

    server.LoadModules("config.txt")

    server.Listen(func(data []byte) []byte {
        fmt.Println("Received data:", string(data))
        return []byte("Hello, client!")
    })
}
```

### Client

```go
package main

import (
    "fmt"
    "github.com/pynezz/pynezzentials/ipc"
)

func main() {
    client := ipc.NewIPCClient("client1", [4]byte{0, 0, 0, 1})   // Identifier must match the server's identifier
    err := client.Connect()
    if err != {
        fmt.Println("Error:", err)
    }

    msg := client.CreateGenericReq([]byte("Hello, server!"), ipc.MSG_MSG, icp.DATA_JSON)

    response, err = client.SendIPCMessage(msg)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Received response:", string(response))
    }
}
```

## License

[LICENSE](LICENSE)
