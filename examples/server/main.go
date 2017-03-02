package main

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
	"github.com/faceless-saint/go-socketcmd"

	"errors"
	"os"
	"os/exec"
)

const EnvSocketPath = "SOCKET_PATH"

var ExampleSocketPath = "example.sock"

func init() {
	// If environment variable is set, override default socket path
	envPath := os.Getenv(EnvSocketPath)
	if envPath != "" {
		ExampleSocketPath = envPath
	}
}

func main() {
	// Parse command line arguments into an exec.Cmd
	if len(os.Args) < 2 {
		panic(errors.New("not enough arguments - you must specify a command"))
	}
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	cmd := exec.Command(os.Args[1], args...)

	// Send command to the socketcmd.Wrapper
	os.Remove(ExampleSocketPath)
	s, err := socketcmd.NewUnix(ExampleSocketPath, cmd)
	if err != nil {
		panic(err)
	}

	// Start the wrapped command - os.Stdin and os.Stdout are connected to the wrapped process
	if err := s.Run(); err != nil {
		panic(err)
	}
}
