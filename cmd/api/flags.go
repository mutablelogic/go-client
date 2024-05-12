package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	*flag.FlagSet

	cmds  []Cmd
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
	ErrHelp = flag.ErrHelp
)

const (
	_ flagType = iota
	Bool
	String
	Duration
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFlags(name string) *Flags {
	flags := new(Flags)
	flags.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)
	flags.names = make(map[string]*Value)
	flags.FlagSet.Usage = func() {}

	// Register global flags
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

func (flags *Flags) Parse(args []string) error {
	// Parse command line
	if err := flags.FlagSet.Parse(args); err != nil {
		if err == flag.ErrHelp {
			flags.PrintUsage()
		}
		return err
	}

	// If there is a version argument, print the version and exit
	if flags.NArg() == 1 && flags.Arg(0) == "version" {
		flags.PrintVersion()
		return ErrHelp
	}

	// Set client options
	opts := []client.ClientOpt{}
	if flags.GetBool("debug") {
		// Append debug option with optional verbose
		opts = append(opts, client.OptTrace(flags.Output(), flags.GetBool("verbose")))
	}

	// Parse the commands
	if flags.NArg() > 0 {
		if cmd := flags.GetCommandSet(flags.Arg(0)); cmd != nil {
			if err := cmd.Parse(flags, opts...); err != nil {
				fmt.Fprintf(os.Stderr, "%v: %v\n", cmd.Name, err)
				return err
			}
		}
	}

	// Return success
	return nil
}

// GetCommandSet returns a command set from a name, or nil if the command set does
// not exist
func (flags *Flags) GetCommandSet(name string) *Cmd {
	for _, cmd := range flags.cmds {
		if cmd.Name == name {
			return &cmd
		}
	}
	return nil
}

// Get returns the value of a flag, and returns true if the flag exists
func (flags *Flags) Get(name string) (string, bool) {
	if value := flags.FlagSet.Lookup(name); value == nil {
		return "", false
	} else {
		return value.Value.String(), true
	}
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
	if flags.NArg() == 1 {
		cmd := flags.GetCommandSet(flags.Arg(0))
		if cmd != nil {
			flags.PrintCommandUsage(cmd)
		}
	} else {
		fmt.Fprintln(w, "Name:", flags.Name())
		fmt.Fprintln(w, "  General command-line interface to API clients")
		fmt.Fprintln(w, "")

		// Command Sets
		fmt.Fprintln(w, "Command Sets:")
		for _, cmd := range flags.cmds {
			fmt.Fprintln(w, "  ", flags.Name(), cmd.Name)
			fmt.Fprintln(w, "    ", cmd.Description)
			fmt.Fprintln(w, "")
		}

		// Help
		fmt.Fprintln(w, "Help:")
		fmt.Fprintln(w, "  ", flags.Name(), "version")
		fmt.Fprintln(w, "    ", "Return the version of the application")
		fmt.Fprintln(w, "")

		// Help for command sets
		for _, cmd := range flags.cmds {
			fmt.Fprintln(w, "  ", flags.Name(), "-help", cmd.Name)
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
	fmt.Fprintln(w, "Name:", flags.Name(), cmd.Name)
	fmt.Fprintln(w, "  ", cmd.Description)
	fmt.Fprintln(w, "")

	// Help for command sets
	fmt.Fprintln(w, "Commands:")
	for _, fn := range cmd.Fn {
		fmt.Fprintln(w, "  ", flags.Name(), cmd.Name, fn.Name)
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
	flags.VisitAll(func(f *flag.Flag) {
		if flags.names[f.Name].cmd == cmd {
			fmt.Fprintf(w, "  -%v: %v\n", f.Name, f.Usage)
		}
	})
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
	if value, exists := flags.Get(name); !exists {
		return ""
	} else {
		return os.ExpandEnv(value)
	}
}
