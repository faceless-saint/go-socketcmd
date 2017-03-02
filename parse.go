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
	DefaultHeader    = "-1:"
	DefaultParseFunc = func(_ []string) string { return DefaultHeader }
	errMissingHeader = fmt.Errorf("missing/invalid socketcmd header")
	headerRegexp     = regexp.MustCompile("^-?[0-9]*:[0-9]*$")
)

// A ParseFunc determines the proper header for a given command.
type ParseFunc func(args []string) string

// Header representation of the given line count and timeout.
func Header(lines, timeout int) string {
	if timeout <= 0 {
		if lines == 0 {
			return ":"
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
		return 0, 0, errMissingHeader
	}
	s := strings.Split(header, ":")
	if len(s) != 2 {
		return 0, 0, errMissingHeader
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
