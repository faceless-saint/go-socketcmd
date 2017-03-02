package socketcmd

/*  Copyright 2017 Ryan Clarke

    This file is part of Socketcmd.

    Socketcmd is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Socketcmd is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Socketcmd.  If not, see <http://www.gnu.org/licenses/>
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

const (
	// Maximum allowed timeout in milliseconds
	//MaxTimeout = 30000
	// Default timeout in milliseconds
	DefaultTimeout = 1000
	// Buffer size in bytes for incoming connections
	ConnBufferSize = 2048
)

/* A Handler manages network socket and I/O redirection.
 */
type Handler interface {
	// Close the socket listener.
	Close() error
	// Start the goroutines for managing process I/O redirection.
	Start()
}

/* NewHandler returns a new Handler for the given socket listener and I/O pipes.
 */
func NewHandler(listener net.Listener, stdin io.Writer, stdout io.Reader) Handler {
	return &handler{listener, stdin, stdout,
		make(chan string, 0),
		make(chan string, 0),
		make(chan bool, 1),
	}
}

type handler struct {
	Socket net.Listener
	Stdin  io.Writer
	Stdout io.Reader

	rch chan string
	wch chan string
	blk chan bool
}

func (h *handler) Close() error {
	return h.Socket.Close()
}

func (h *handler) Start() {
	log.Println("Starting socket handler")
	go h.HandleSocket()
	go h.HandleStdin()
	go h.ListenStdin()
	go h.ListenStdout()
}

/* goroutine: forward socket connections to the wrapped process
 *		socket -> w_chan, r_chan -> socket
 */
func (h *handler) HandleSocket() {
	for {
		conn, err := h.Socket.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Println(err)
			continue
		}
		if err := h.handleConnection(conn); err != nil {
			log.Println(err)
		}
	}
}

func (h *handler) handleConnection(conn net.Conn) error {
	defer conn.Close() // close the connection when finished

	// Read the command from the socket connection
	buf := make([]byte, ConnBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	// [lines]:[timeout] args...
	words := strings.SplitN(string(buf[:n]), " ", 2)
	if len(words) < 2 {
		words = append(words, "")
	}

	// Parse header word for line count and timeout information
	lines, timeout, err := ParseHeader(words[0])
	if err != nil {
		_, err2 := io.WriteString(conn, err.Error())
		return err2
	}

	// Block the response consumer while handling the connection
	h.blk <- true
	defer func() { h.blk <- false }()

	// Send command to the stdin Writer
	log.Printf("(%s)-> %s\n", conn.RemoteAddr().String(), words[1])
	h.wch <- words[1]

	// Send the captured response to the socket connection
	return sendResponse(conn, h.rch, lines, timeout)
}

func sendResponse(conn net.Conn, resp <-chan string, lines, timeout int) error {
	var count int
	// Use default timeout if given value is out of bounds
	if timeout <= 0 {
		timeout = DefaultTimeout
	} /*else if timeout > MaxTimeout {	//TODO: enable maximum timeout?
		timeout = MaxTimeout
	}*/
	for {
		// Skip line counting if lines is negative
		if lines >= 0 && count >= lines {
			return nil
		}
		t := time.NewTimer(time.Duration(timeout) * time.Millisecond)
		select {
		case line, ok := <-resp:
			if !t.Stop() {
				<-t.C
			}
			if !ok {
				return nil
			}
			// Send response line to socket connection
			if _, err := io.WriteString(conn, line+"\n"); err != nil {
				return err
			}
			if lines >= 0 {
				count++
			}
		case <-t.C:
			// Timeout exceeded
			return nil
		}
	}
}

/* goroutine: forward writes from the write channel to cmd.Stdin
 *		w_chan -> cmd.Stdin
 */
func (h *handler) HandleStdin() {
	for {
		input, ok := <-h.wch
		if !ok {
			return
		}
		if _, err := io.WriteString(h.Stdin, input+"\n"); err != nil {
			log.Println(err)
		}
	}
}

/* goroutine: forward writes from os.Stdin to the write channel
 *		os.Stdin -> w_chan
 */
func (h *handler) ListenStdin() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		h.wch <- scanner.Text()
	}
	if scanner.Err() != nil {
		log.Println(scanner.Err())
	}
}

/* goroutine: forward reads of cmd.Stdout to os.Stdout and the read channel
 *		cmd.Stdout -> os.Stdout + r_chan
 */
func (h *handler) ListenStdout() {
	go h.consumeStdout()
	scanner := bufio.NewScanner(h.Stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		h.rch <- scanner.Text()
	}
	if scanner.Err() != nil {
		log.Println(scanner.Err())
	}
	close(h.rch)
}

/* goroutine: keep read channel empty when no socket connection is present
 *		r_chan -> /dev/null (iff no socket connection)
 */
func (h *handler) consumeStdout() {
	for {
		select {
		case _, ok := <-h.rch:
			if !ok {
				return
			}
		case <-h.blk:
			<-h.blk
		}
	}
}
