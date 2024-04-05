# Pynezzentials

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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
