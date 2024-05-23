package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/version"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	*flag.FlagSet

	cmds  []Cmd
	cmd   *Cmd
	root  string
	fn    string
	args  []string
	names map[string]*Value
}

type Value struct {
	flagType
	cmd string
}

type flagType uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	ErrHelp    = flag.ErrHelp
	ErrInstall = errors.New("install")
)

const (
	_ flagType = iota
	Bool
	String
	Duration
	Float
	Unsigned
)

var (
	reExt = regexp.MustCompile(`^[a-zA-Z0-9]{1,32}$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFlags(name string) *Flags {
	flags := new(Flags)
	flags.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)
	flags.names = make(map[string]*Value)
	flags.FlagSet.Usage = func() {}

	// Register global flags
	flags.String("", "out", "", "Set output filename or type")
	flags.Bool("", "debug", false, "Enable debug logging")
	flags.Bool("", "verbose", false, "Enable verbose output")
	flags.Duration("", "timeout", 0, "Client timeout")

	// Return success
	return flags
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (flags *Flags) Register(c Cmd) {
	flags.cmds = append(flags.cmds, c)
}

func (flags *Flags) Parse(args []string) (*Fn, []string, error) {
	// Parse command line
	err := flags.FlagSet.Parse(args)

	// Check for global commands
	if flags.NArg() == 1 {
		switch flags.Arg(0) {
		case "version":
			flags.PrintVersion()
			return nil, nil, ErrHelp
		case "help":
			flags.PrintUsage()
			return nil, nil, ErrHelp
		case ErrInstall.Error():
			return nil, nil, ErrInstall
		}
	}

	if cmd := flags.getCommandSet(flags.Name()); cmd != nil {
		// If the name of the command is the same as the name of the application
		flags.cmd = cmd
		flags.root = cmd.Name
		if len(flags.Args()) > 0 {
			flags.fn = flags.Arg(0)
			if len(flags.Args()) > 1 {
				flags.args = flags.Args()[1:]
			}
		}
	} else if flags.NArg() > 0 {
		if cmd := flags.getCommandSet(flags.Arg(0)); cmd != nil {
			flags.cmd = cmd
			flags.root = strings.Join([]string{flags.Name(), cmd.Name}, " ")
			flags.fn = flags.Arg(1)
			if len(flags.Args()) > 1 {
				flags.args = flags.Args()[2:]
			}
		}
	}

	if flags.GetBool("debug") {
		fmt.Fprintf(os.Stderr, "Function: %q Args: %q\n", flags.fn, flags.args)
	}

	// Print usage
	if err != nil {
		if err != flag.ErrHelp {
			// TODO: Do nothing
		} else {
			flags.PrintUsage()
		}
		return nil, nil, err
	} else if flags.cmd == nil {
		fmt.Fprintf(os.Stderr, "Unknown command, try \"%s -help\"\n", flags.Name())
		return nil, nil, ErrHelp
	}

	// Set client options
	opts := []client.ClientOpt{}
	if flags.GetBool("debug") {
		// Append debug option with optional verbose
		opts = append(opts, client.OptTrace(flags.Output(), flags.GetBool("verbose")))
	}

	// If there is a help argument, print the help and exit
	if flags.NArg() == 1 && flags.Arg(0) == "help" || flags.cmd == nil {
		flags.PrintUsage()
		return nil, nil, ErrHelp
	} else if err := flags.cmd.Parse(flags, opts...); err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", flags.cmd.Name, err)
		return nil, nil, err
	}

	// Set the function to call
	fn := flags.cmd.Get(flags.fn)
	if fn == nil {
		fmt.Fprintf(os.Stderr, "Unknown command, try \"%s -help\"\n", flags.Name())
		return nil, nil, ErrHelp
	}

	// Check the number of arguments
	if err := fn.CheckArgs(flags.args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	// Return success
	return fn, flags.args, nil
}

// Get returns the value of a flag, and returns true if the flag exists
// and has been changed from the default
func (flags *Flags) Get(name string) (string, bool) {
	var visited bool
	flags.Visit(func(f *flag.Flag) {
		if f.Name == name {
			visited = true
		}
	})
	if value := flags.FlagSet.Lookup(name); value == nil {
		return "", false
	} else {
		return value.Value.String(), visited
	}
}

// Get the current command set
func (flags *Flags) Cmd() *Cmd {
	return flags.cmd
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// PrintVersion prints the version of the application
func (flags *Flags) PrintVersion() {
	w := flags.Output()
	fmt.Fprintf(w, "%v", flags.Name())
	if version.GitSource != "" {
		if version.GitTag != "" {
			fmt.Fprintf(w, " %v", version.GitTag)
		}
		if version.GitSource != "" {
			fmt.Fprintf(w, " (%v)", version.GitSource)
		}
		fmt.Fprintln(w, "")
	}
	if runtime.Version() != "" {
		fmt.Fprintf(w, "%v %v/%v\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	}
	if version.GitBranch != "" {
		fmt.Fprintf(w, "Branch: %v\n", version.GitBranch)
	}
	if version.GitHash != "" {
		fmt.Fprintf(w, "Hash: %v\n", version.GitHash)
	}
	if version.GoBuildTime != "" {
		fmt.Fprintf(w, "BuildTime: %v\n", version.GoBuildTime)
	}
}

// PrintUsage prints the usage of the application
func (flags *Flags) PrintUsage() {
	w := flags.Output()
	if flags.cmd != nil {
		flags.PrintCommandUsage(flags.cmd)
	} else {
		fmt.Fprintln(w, "Name:", flags.root)
		fmt.Fprintln(w, "  General command-line interface to API clients")
		fmt.Fprintln(w, "")

		// Command Sets
		fmt.Fprintln(w, "Command Sets:")
		for _, cmd := range flags.cmds {
			fmt.Fprintln(w, "  ", flags.root, cmd.Name)
			fmt.Fprintln(w, "    ", cmd.Description)
			fmt.Fprintln(w, "")
		}

		// Help
		fmt.Fprintln(w, "Help:")
		fmt.Fprintln(w, "  ", flags.root, "version")
		fmt.Fprintln(w, "    ", "Return the version of the application")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  ", flags.root, "install")
		fmt.Fprintln(w, "    ", "Install symlinks for command calling")
		fmt.Fprintln(w, "")

		// Help for command sets
		for _, cmd := range flags.cmds {
			fmt.Fprintln(w, "  ", flags.root, "-help", cmd.Name)
			fmt.Fprintln(w, "    ", "Display", cmd.Name, "command syntax")
			fmt.Fprintln(w, "")
		}

		fmt.Fprintln(w, "")
		flags.PrintGlobalFlags()
	}
}

// PrintGlobalFlags prints out the global flags
func (flags *Flags) PrintGlobalFlags() {
	flags.PrintCommandFlags("")
}

// PrintCommandUsage prints the usage of a commandset
func (flags *Flags) PrintCommandUsage(cmd *Cmd) {
	w := flags.Output()
	fmt.Fprintln(w, "Name:", flags.root)
	fmt.Fprintln(w, "  ", cmd.Description)
	fmt.Fprintln(w, "")

	// Help for command sets
	fmt.Fprintln(w, "Commands:")
	for _, fn := range cmd.Fn {
		fmt.Fprintln(w, "  ", flags.root, fn.Name, fn.Syntax)
		fmt.Fprintln(w, "    ", fn.Description)
		fmt.Fprintln(w, "")
	}
	flags.PrintCommandFlags(cmd.Name)
	fmt.Fprintln(w, "")
	flags.PrintGlobalFlags()
}

// PrintGlobalFlags prints out the global flags
func (flags *Flags) PrintCommandFlags(cmd string) {
	w := flags.Output()
	if cmd == "" {
		fmt.Fprintln(w, "Global flags:")
	} else {
		fmt.Fprintf(w, "Flags for %v:\n", cmd)
	}
	flags.VisitAll(func(flag *flag.Flag) {
		if flags.names[flag.Name].cmd == cmd {
			printFlag(w, flag)
		}
	})
}

func printFlag(w io.Writer, f *flag.Flag) {
	fmt.Fprintf(w, "  -%v", f.Name)
	if len(f.DefValue) > 0 {
		fmt.Fprintf(w, " (default %q)", f.DefValue)
	}
	fmt.Fprintf(w, "\n    %v\n\n", f.Usage)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - SET FLAGS

func (flags *Flags) Bool(cmd, name string, value bool, usage string) *Value {
	// TODO: Panic if the flag already exists
	if _, ok := flags.names[name]; ok {
		panic(fmt.Sprintf("flag redefined: %q", name))
	}

	// Create the flag
	result := &Value{
		cmd:      cmd,
		flagType: Bool,
	}

	// Set up flag
	flags.FlagSet.Bool(name, value, usage)
	flags.names[name] = result

	// Return success
	return result
}

func (flags *Flags) String(cmd, name string, value string, usage string) *Value {
	// TODO: Panic if the flag already exists
	if _, ok := flags.names[name]; ok {
		panic(fmt.Sprintf("flag redefined: %q", name))
	}

	// Create the flag
	result := &Value{
		cmd:      cmd,
		flagType: String,
	}

	// Set up flag
	flags.FlagSet.String(name, value, usage)
	flags.names[name] = result

	// Return success
	return result
}

func (flags *Flags) Duration(cmd, name string, value time.Duration, usage string) *Value {
	// TODO: Panic if the flag already exists
	if _, ok := flags.names[name]; ok {
		panic(fmt.Sprintf("flag redefined: %q", name))
	}

	// Create the flag
	result := &Value{
		cmd:      cmd,
		flagType: Duration,
	}

	// Set up flag
	flags.FlagSet.Duration(name, value, usage)
	flags.names[name] = result

	// Return success
	return result
}

func (flags *Flags) Float(cmd, name string, value float64, usage string) *Value {
	// TODO: Panic if the flag already exists
	if _, ok := flags.names[name]; ok {
		panic(fmt.Sprintf("flag redefined: %q", name))
	}

	// Create the flag
	result := &Value{
		cmd:      cmd,
		flagType: Float,
	}

	// Set up flag
	flags.FlagSet.Float64(name, value, usage)
	flags.names[name] = result

	// Return success
	return result
}

func (flags *Flags) Unsigned(cmd, name string, value uint64, usage string) *Value {
	// TODO: Panic if the flag already exists
	if _, ok := flags.names[name]; ok {
		panic(fmt.Sprintf("flag redefined: %q", name))
	}

	// Create the flag
	result := &Value{
		cmd:      cmd,
		flagType: Unsigned,
	}

	// Set up flag
	flags.FlagSet.Uint64(name, value, usage)
	flags.names[name] = result

	// Return success
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - GET FLAGS

// GetBool returns a flag value of type bool
func (flags *Flags) GetBool(name string) bool {
	if value := flags.FlagSet.Lookup(name); value == nil {
		return false
	} else {
		return value.Value.(flag.Getter).Get().(bool)
	}
}

// GetString returns a flag value as a string, and expands
// any environment variables in the returned value
func (flags *Flags) GetString(name string) string {
	value, _ := flags.Get(name)
	return os.ExpandEnv(value)
}

// GetValue returns a flag value, if it has been changed from the default
// value
func (flags *Flags) GetValue(name string) (any, error) {
	value, visited := flags.Get(name)
	if !visited {
		return nil, ErrNotFound.With(name)
	} else {
		value = os.ExpandEnv(value)
	}
	switch flags.names[name].flagType {
	case Bool:
		return strconv.ParseBool(value)
	case String:
		return value, nil
	case Duration:
		return time.ParseDuration(value)
	case Float:
		return strconv.ParseFloat(value, 64)
	case Unsigned:
		return strconv.ParseUint(value, 10, 64)
	default:
		return nil, ErrNotImplemented.With(flags.names[name].flagType)
	}
}

// GetOutExt returns the extension of the output file, or an empty string
// if the output file is not set. It does not include the initial '.' character
func (flags *Flags) GetOutExt() string {
	value, exists := flags.Get("out")
	if !exists {
		return ""
	} else if ext := filepath.Ext(value); ext == "" && reExt.MatchString(value) {
		return value
	} else if len(ext) > 1 && ext[0] == '.' && reExt.MatchString(ext[1:]) {
		return ext[1:]
	} else {
		return ""
	}
}

// GetOutPath returns the full pathname of the output file, or an empty string
// if the output file is not set.
func (flags *Flags) GetOutPath() string {
	value, exists := flags.Get("out")
	if !exists {
		return ""
	} else if flags.GetOutExt() == value {
		// -output csv or something, without a file extension
		return ""
	} else {
		return value
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (flags *Flags) getCommandSet(name string) *Cmd {
	for _, cmd := range flags.cmds {
		if cmd.Name == name {
			return &cmd
		}
	}
	return nil
}
