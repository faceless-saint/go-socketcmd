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

// DefaultTimeout is the timeout used when the parsed timeout is missing or invalid.
var DefaultTimeout = 5000

/* A ParseFunc indicates the number of lines expected to be written to stdout as a result of the
 * given command. It also indicates a timeout for the interval between lines.
 */
type ParseFunc func(cmd string) (lines, timeout int)

// A ParseOpt represents a parsing result for a particular command, including line count and timeout.
type ParseOpt struct {
	Lines int
	Time  int
}

// NewParseFunc returns a new ParseFunc using the given map of parsing options and a default value.
func NewParseFunc(opts map[string]ParseOpt, def ParseOpt) ParseFunc {
	return func(cmd string) (lines, timeout int) {
		if opt, ok := opts[cmd]; ok {
			return opt.Lines, opt.Time
		}
		return def.Lines, def.Time
	}
}

// TimeoutOnlyParse returns a ParseFunc with a universal timeout for all commands
func TimeoutOnlyParse(timeout int) ParseFunc {
	return func(_ string) (lines, timeout int) {
		return -1, timeout
	}
}
