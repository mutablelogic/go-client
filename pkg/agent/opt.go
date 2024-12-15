package agent

import "fmt"

//////////////////////////////////////////////////////////////////
// TYPES

type Opts struct {
	Tools    []Tool
	StreamFn func(Response)
}

type Opt func(*Opts) error

//////////////////////////////////////////////////////////////////
// METHODS

// OptStream sets the stream function, which is called during the
// response generation process
func OptStream(fn func(Response)) Opt {
	return func(o *Opts) error {
		o.StreamFn = fn
		return nil
	}
}

// OptTools sets the tools for the chat request
func OptTools(t ...Tool) Opt {
	return func(o *Opts) error {
		if len(t) == 0 {
			return fmt.Errorf("no tools specified")
		}
		for _, tool := range t {
			if tool == nil {
				return fmt.Errorf("nil tool specified")
			}
			o.Tools = append(o.Tools, tool)
		}
		return nil
	}
}
