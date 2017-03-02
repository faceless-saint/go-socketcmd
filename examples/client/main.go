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

	"os"
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
	// Connect to the socket with a new client using the default parsing function
	client := socketcmd.NewClient("unix", ExampleSocketPath, nil)

	// Send command line arguments to the socket
	if err := client.Send(os.Args[1:]...); err != nil {
		panic(err)
	}
}
