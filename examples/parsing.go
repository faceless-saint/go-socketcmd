package examples

import "github.com/faceless-saint/go-socketcmd"

// sets command headers for top-level commands
var ExampleBaseCommandTable = map[string]string{
	// at most 2 lines expected, default timeout
	"command": socketcmd.Header(2, 0),
	// defualt header = unlimited lines, default timeout
	"base": socketcmd.DefaultHeader,
	// empty header = do not wait for a response
	"blind": socketcmd.EmptyHeader,
}

// sets nested command headers for subcommands
var ExampleSubCommandTable = map[string]map[string]string{
	// will be associated with the "base" command above
	"base": map[string]string{
		"sub1": socketcmd.DefaultHeader,
		"sub2": socketcmd.DefaultHeader,
	},
}

// blacklist - defined commands that are forbidden
var ExampleCommandBlacklist = map[string]string{
	"bad-command": socketcmd.ForbiddenHeader,
	"forbidden":   socketcmd.ForbiddenHeader,
	// set a custom header with 10s timeout - this command is allowed
	"allowed": socketcmd.Header(-1, 10000),
}

// set up a whitelist by using ForbiddenHeader as the default and then
// providing a list of recognized commands
func GetWhitelistParseFunc() ParseFunc {
	args := socketcmd.NewArguments(ExampleBaseCommandTable, socketcmd.ForbiddenHeader)
	args.AddNestedArguments(ExampleSubCommandTable)
	return args.ParseFunc()
}

// set up a blacklist by using a list of commands with the headers set to
// ForbiddenHeader, and a standard default header
func GetBlacklistParseFunc() ParseFunc {
	args := socketcmd.NewArguments(ExampleCommandBlacklist, socketcmd.DefaultHeader)
	return args.ParseFunc()
}
