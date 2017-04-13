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
	"context"
	"io"
	"net"
	"strings"
)

// A Client connects to a Wrapper's socket.
type Client interface {
	/* Dialer sets the dialer configuration used by the Client.
	 */
	Dialer(net.Dialer)
	/* Send the given arguments to the socket Wrapper. The Client's parser function is used
	 * to generate the socketcmd header appropriate for the given arguments.
	 */
	Send(args ...string) ([]string, error)
	/* SendContext sends the given arguments to the socket Wrapper, using the given context
	 * to manage the connection. The Client's parser function is used to generate the
	 * socketcmd header appropriate for the given arguments.
	 */
	SendContext(ctx context.Context, args ...string) ([]string, error)
}

// NewClient returns a new Client for the given socket address and parser.
func NewClient(proto, addr string, parser ParseFunc) Client {
	if parser == nil {
		parser = DefaultParseFunc
	}
	return &client{parser, proto, addr, net.Dialer{}}
}

type client struct {
	Parse    ParseFunc
	Protocol string
	Address  string

	d net.Dialer
}

func (c *client) Dialer(dialer net.Dialer) {
	c.d = dialer
}

func (c *client) Send(args ...string) ([]string, error) {
	header := c.Parse(args)
	if header == ForbiddenHeader {
		return nil, ErrCommandForbidden
	}

	conn, err := c.d.Dial(c.Protocol, c.Address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return c.send(conn, header, args...)
}

func (c *client) SendContext(ctx context.Context, args ...string) ([]string, error) {
	header := c.Parse(args)
	if header == ForbiddenHeader {
		return nil, ErrCommandForbidden
	}

	conn, err := c.d.DialContext(ctx, c.Protocol, c.Address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return c.send(conn, header, args...)
}

func (c *client) send(conn net.Conn, header string, args ...string) ([]string, error) {
	// Send command and get response scanner
	scanner, err := c.stream(conn, header, args...)
	if err != nil {
		return nil, err
	}

	// Print the socket responses until the connection is closed
	var results []string
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}
	return results, scanner.Err()
}

func (c *client) stream(conn net.Conn, header string, args ...string) (
	*bufio.Scanner, error,
) {
	// Send the given arguments to the socket as a space-separated string
	if _, err := io.WriteString(conn, header+" "+strings.Join(args, " ")); err != nil {
		return nil, err
	}
	return bufio.NewScanner(conn), nil
}
