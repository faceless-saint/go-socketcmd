# go-socketcmd

A Golang library to enable socket-based communication with applications not designed to support it. This allows for communicating with the wrapped application's stdin through the socket, as well as configurable stdout mirroring back along the socket. While running, the wrapped process will also be attached to the caller's stdin and stdout, allowing for transparent interaction in addition to the socket integration.

This package is primarily designed with Unix domain sockets in mind, though any standard implementation of net.Listener should be compatible with the wrapper, including TCP sockets. See the [offiical documentation](https://golang.org/pkg/net/#Listener) for more details.

**Note:** Only one connection will be processed at a time to prevent data corruption and race conditions. Multiple clients may connect to the socket, but those connections will be accepted and processed serially.

## Wrapper Usage

Browse the provided examples to get started.

#### Package installation
`go get github.com/faceless-saint/go-socketcmd`

#### Setup and wrapper execution
```go
import (
	"github.com/faceless-saint/go-socketcmd"

	"net"
	"os/exec"
)

// Note: error handling has been omitted for brevity

func main() {
	var cmd *exec.Cmd
	//... prepare the exec.Cmd object here (see examples)

	// Create listener for the socketcmd.Wrapper
	ln, err := net.Listen("tcp", "0.0.0.0")

	// create wrapper using the given net.Listener
	tcpWrapper, err := socketcmd.New(ln, cmd)

	// create wrapper using the specified UNIX socket
	unixWrapper, err := socketcmd.NewUnix("/path/to/control.sock", cmd)

	// Start the wrapped command
	err = tcpWrapper.Start()

	// Wait for the wrapped command to exit
	err = tcpWrapper.Wait()

	// Run the wrapped command and block until it exits
	err = unixWrapper.Run()
}
```

#### Client connections
```go
import (
	"github.com/faceless-saint/go-socketcmd"

	"context"
	"net"
	"os/exec"
)

// Note: error handling has been omitted for brevity

func main() {
	// Command arguments to send - will be joined into a space-separated string
	command := []string{"cmd", "arg1", "arg2"}

	// Create a new Client connecting to the given socket using the default header
	client := socketcmd.NewClient("unix", "/path/to/control.sock", nil)

	// Create a new Client with a custom parse function
	parsef = func(args []string) string {
		return "-1:" // Placeholder - return the default header
	}
	customClient := socketcmd.NewClient("unix", "/path/to/control.sock", parsef)

	// Send a command to the socket. The header is generated using the client's parse function
	err := client.Send(command...)

	// Command arguments to send, including the socketcmd header
	command = []string{"-1:", "cmd", "arg1", "arg2"}

	// Send a command, manually specifying the desired socktcmd header
	err = client.Send(command...)

	// You can also provide a context for the connection
	err = client.SendContext(context.Background, command...)
}
```
#### Header parsing rules
The socketcmd header is in the following format: `[n]:[t]`

General rules:
* Both `n` and `t` must be integers
* Either value (but not the `:`) may be omitted to use the default value

Lines:
* If `n` is less than 1, then an unlimited number of lines may be read
* If `n` is exactly 0, then no lines will be read regardless of timeout
* If `n` is greater than 0, then at most `n` lines will be read

Timeout:
* `t` cannot be negative
* If `t` is 0, then the default timeout will be used
* If `t` is greater than 0, then the timeout will be `t` milliseconds

*Setting a maximum line count is very useful if you know exactly how many lines of stdout will result from the send command, because this allows the client to return immediately after receiving those lines. Otherwise, you must wait for the timeout to elapse before releasing the connection.*

Examples:
* `-1:` - default header, read unlimited lines until the default timeout has elapsed
* `:` - null header, expect no lines of output and exit immediately
* `4:` - counting header, read at most 4 lines of output with the default timeout
* `-1:10000` - waiting header, like default but with a custom (typically slower) timeout
* `4:10000` - combination of counting and waiting, read at most 4 lines with a custom timeout
