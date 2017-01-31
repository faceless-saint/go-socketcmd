package socketcmd
/*	Copyright 2017 Ryan Clarke

	This file is part of socketcmd.

	socketcmd is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Foobar is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with socketcmd.  If not, see <http://www.gnu.org/licenses/>
*/

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var errLog = log.New(os.Stderr, "", log.LstdFlags)

/* A Handler manages the socket interface for a Wrapper. It captures stdout from the wrapped process and
 * passes commands from incoming connections to the wrapped stdin.
 */
type Handler interface {
	/* CaptureStdout starts a long-running scanner on the stdout reader. Each line scanned is echoed to
	 * stdout. If blocking mode is enabled, each line is also sent to the `resp` channel between scans.
	 * Blocking mode is configured by reading from the `block` channel.
	 */
	CaptureStdout()

	/* HandleSocket starts a long-running handler for the socket listener. Input to the socket is passed
	 * to the stdin writer, and the response is returned. Sets the scanner created by the `CaptureStdout`
	 * method to blocking mode while connections are being processed and reads the corresponding response
	 * lines from the `resp` channel.
	 */
	HandleSocket()

	// Close the socket listener for the Handler. Causes the goroutine started by `HandleSocket` to exit.
	Close() error
}

/* NewHandler returns a new Handler for the given socket and I/O streams. The given parsing function is
 * used to determine the timeout and number of response lines for each received command. If this parsing
 * function is undefined then it defaults to a timeout of 2 seconds.
 */
func NewHandler(sock net.Listener, stdin io.Writer, stdout io.Reader, parsef ParseFunc) Handler {
	if parsef == nil {
		parsef = TimeoutOnlyParse(DefaultTimeout)
	}
	return &handler{parsef, sock, stdin, stdout, make(chan bool, 0), make(chan string, 0)}
}

type handler struct {
	Parse  ParseFunc
	Socket net.Listener
	Stdin  io.Writer
	Stdout io.Reader

	block chan bool
	resp  chan string
}

func (h *handler) CaptureStdout() {
	var blocking bool

	// Scan each line of the piped stdout
	scanner := bufio.NewScanner(h.Stdout)
	for scanner.Scan() {
		// Echo to stdout of main process
		fmt.Println(scanner.Text())

		select {
		default:
		case b := <-h.block:
			// Update blocking mode if a signal is received
			blocking = b
		}
		// If blocking mode is set, send stdout to the channel between scans
		if blocking {
			h.resp <- scanner.Text()
		}
	}
	if scanner.Err() != nil {
		errLog.Println(scanner.Err())
	}
	// Close the response channel when stdout closes
	close(h.resp)
}

func (h *handler) HandleSocket() {
	for {
		conn, err := h.Socket.Accept()
		if err != nil {
			// Exit cleanly if the error is a closed listener
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			errLog.Println(err)
			continue
		}
		if err := h.handleConnection(conn); err != nil {
			errLog.Println(err)
		}
	}
}

func (h *handler) handleConnection(conn net.Conn) error {
	defer conn.Close()
	// Read the command from the socket connection
	var buf []byte
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}

	// Parse the command to determine response lines and timeout
	lines, timeout := h.Parse(string(buf))

	// Ensure that the command ends with a newline
	if strings.HasSuffix(string(buf), "\n") {
		buf = append(buf, "\n"...)
	}
	// Echo command back to the socket connection
	if _, err := conn.Write(buf); err != nil {
		return err
	}

	// Block the stdout scanner until the handling is complete
	h.block <- true
	defer func() { h.block <- false }()

	// Send command to the stdin Writer
	if _, err := h.Stdin.Write(buf); err != nil {
		io.WriteString(conn, "error: "+err.Error()+"\n")
		return err
	}

	// Send the captured response to the socket connection
	_, err := io.WriteString(conn, getResponse(h.resp, lines, timeout))
	return err
}

func getResponse(resp <-chan string, lines int, timeout int) string {
	var (
		response string
		count    int
	)
	// Use default timeout for package if given value is out of bounds
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	for {
		// Count response lines if lines >= 0
		if lines >= 0 && count >= lines {
			return response
		}
		t := time.NewTimer(time.Duration(timeout) * time.Millisecond)
		select {
		case line, ok := <-resp:
			if !t.Stop() {
				<-t.C
			}
			if !ok {
				// Response channel has closed
				return response
			}
			// Add new line to the response
			response += line + "\n"
			if lines >= 0 {
				count++
			}
		case <-t.C:
			// Timeout exceeded since last response
			return response
		}
	}
}

func (h *handler) Close() error {
	return h.Socket.Close()
}
