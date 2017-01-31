# socketcmd

A wrapper library to enable socket-based communication with applications not designed to support
it. This allows for communicating with the wrapped application's stdin through the socket, as well
as configurable stdout mirroring back along the socket.

This package is primarily designed with Unix domain sockets in mind, though any
implementation of net.Listener should be compatible with the wrapper.

**Note: This package is not meant for concurrent socket access. Multiple connections
to the socket are permitted, but only one will be processed at a time to prevent corruption
of the data stream. The blocking nature of the listener is deliberate.**

## Wrapper Usage
`TODO`
