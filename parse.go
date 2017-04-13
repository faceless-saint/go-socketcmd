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
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Read unlimited lines until the default timeout
	DefaultHeader = "-1:"

	// Command not allowed (will not be sent to Wrapper)
	ForbiddenHeader = "-:"

	// Do not wait for any response
	EmptyHeader = ":"

	headerRegexp = regexp.MustCompile("^-?[0-9]*:[0-9]*$")

	ErrMissingHeader    = fmt.Errorf("missing or invalid socketcmd header")
	ErrCommandForbidden = fmt.Errorf("the provided command is not allowed")
)

// A ParseFunc determines the proper header for a given command sequence.
type ParseFunc func(args []string) string

// Header representation of the given line count and timeout.
func Header(lines, timeout int) string {
	if timeout <= 0 {
		if lines == 0 {
			return EmptyHeader
		}
		return fmt.Sprintf("%d:", lines)
	}
	if lines == 0 {
		return fmt.Sprintf(":%d", timeout)
	}
	return fmt.Sprintf("%d:%d", lines, timeout)
}

// ParseHeader extracts the line count and timeout from the given header.
func ParseHeader(header string) (lines, timeout int, err error) {
	if !headerRegexp.MatchString(header) {
		return 0, 0, ErrMissingHeader
	}
	s := strings.Split(header, ":")
	if len(s) != 2 {
		return 0, 0, ErrMissingHeader
	}
	if s[0] != "" {
		lines, err = strconv.Atoi(s[0])
		if err != nil {
			return
		}
	}
	if s[1] != "" {
		timeout, err = strconv.Atoi(s[1])
	}
	return
}

func DefaultParseFunc(_ []string) string {
	return DefaultHeader
}

func NewArguments(table map[string]string, defaultHeader string) Argument {
	a := Argument{make(map[string]Argument, len(table)), defaultHeader}
	for cmd, header := range table {
		a.Args[cmd] = Argument{nil, header}
	}
	return a
}

type Argument struct {
	Args   map[string]Argument
	Header string
}

func (a *Argument) Match(args []string) string {
	// Use default if header is omitted
	if a.Header == "" {
		a.Header = DefaultHeader
	}
	if len(args) < 1 {
		return a.Header
	}
	if arg, ok := a.Args[args[0]]; ok {
		// Propagate headers to child elements
		if arg.Header == "" {
			arg.Header = a.Header
		}
		if len(args) < 2 {
			//return arg.Match([]string{})
		}
		return arg.Match(args[1:])
	}
	return a.Header
}

func (a *Argument) AddArguments(table map[string]string) {
	if a.Args == nil {
		a.Args = make(map[string]Argument, len(table))
	}
	for arg, header := range table {
		if header == "" {
			header = a.Header
		}
		a.Args[arg] = Argument{nil, header}
	}
}

func (a *Argument) AddNestedArguments(table map[string]map[string]string) {
	if a.Args == nil {
		a.Args = make(map[string]Argument, len(table))
	}
	for cmd, args := range table {
		if arg, ok := a.Args[cmd]; ok {
			arg.AddArguments(args)
			continue
		}
		a.Args[cmd] = NewArguments(args, a.Header)
	}
}

func (a *Argument) ParseFunc() ParseFunc {
	return func(args []string) string {
		return a.Match(args)
	}
}
