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
	"errors"
	"net"
	"os"
	"os/exec"
)

/* A Wrapper provides I/O redirection for a process. Input to the Wrapper's network socket will
 * be forwarded to the process, with the resulting lines of stdout returned to the socket client.
 */
type Wrapper interface {
	// Run the wrapped process.
	Run() error
	// Start the wrapped process.
	Start() error
	// Wait for the wrapped process to exit.
	Wait() error
}

/* NewUnix returns a new socket Wrapper around the given command using a new UNIX domain
 * socket with the given address.
 */
func NewUnix(socket string, cmd *exec.Cmd) (Wrapper, error) {
	os.Remove(socket)
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return nil, err
	}
	return New(listener, cmd)
}

/* New returns a new socket Wrapper around the given command using the given socket listener.
 */
func New(listener net.Listener, cmd *exec.Cmd) (Wrapper, error) {
	if listener == nil || cmd == nil {
		return nil, errors.New("missing required parameters")
	}
	// Create pipes for I/O redirection
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// Initialize socket Handler for the wrapped process
	return &wrapper{cmd, NewHandler(listener, stdin, stdout)}, nil
}

type wrapper struct {
	Cmd *exec.Cmd
	h   Handler
}

func (w *wrapper) Run() error {
	w.h.Start()
	defer w.h.Close()
	return w.Cmd.Run()
}

func (w *wrapper) Start() error {
	w.h.Start()
	return w.Cmd.Start()
}

func (w *wrapper) Wait() error {
	defer w.h.Close()
	return w.Cmd.Wait()
}
