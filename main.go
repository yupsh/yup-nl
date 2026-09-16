// Command yup-nl is the CLI wrapper around github.com/gloo-foo/cmd-nl.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-nl"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name          = "nl"
	flagBody      = "body-numbering"
	flagSep       = "number-separator"
	flagStart     = "starting-line-number"
	flagIncrement = "line-increment"
	flagWidth     = "number-width"
	flagFormat    = "number-format"
)

// Error is the package sentinel error type. Every error this package emits is
// declared as a const of this type so callers can match with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

const (
	// ErrInvalidBody is returned for an unrecognized -b style.
	ErrInvalidBody Error = "invalid body-numbering style (want a, t, or n)"
	// ErrInvalidFormat is returned for an unrecognized -n format.
	ErrInvalidFormat Error = "invalid number-format (want ln, rn, or rz)"
)

// bodyOptions maps each supported -b style to its library option constant.
var bodyOptions = map[string]any{
	"a": command.NlBodyAll,
	"t": command.NlBodyNonEmpty,
	"n": command.NlBodyNone,
}

// formatValues is the set of supported -n format strings.
var formatValues = map[string]struct{}{
	"ln": {}, "rn": {}, "rz": {},
}

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `nl [OPTIONS] [FILE...]

Write each FILE to standard output, with line numbers added.
With no FILE, or when FILE is -, read standard input.`

// spec declares the nl wrapper: a file-or-stdin filter with numbering flags.
var spec = clix.Spec{
	Name:     name,
	Summary:  "number lines of files",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// flags builds a fresh set of the wrapper's flags. It is a constructor rather
// than a shared slice so each parse gets its own flag instances (urfave/cli
// retains a flag's set-state on its pointer across runs, which would otherwise
// leak between test invocations).
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.StringFlag{
			Name:    flagBody,
			Aliases: []string{"b"},
			Usage:   "use STYLE for numbering body lines (a=all, t=non-empty, n=none)",
			Sources: urf.EnvVars("YUP_NL_BODY_NUMBERING"),
			Value:   "",
		},
		&urf.StringFlag{
			Name:    flagSep,
			Aliases: []string{"s"},
			Usage:   "add STRING after (possible) line number",
			Sources: urf.EnvVars("YUP_NL_NUMBER_SEPARATOR"),
			Value:   "",
		},
		&urf.IntFlag{
			Name:    flagStart,
			Aliases: []string{"v"},
			Usage:   "first line number for each section",
			Sources: urf.EnvVars("YUP_NL_STARTING_LINE_NUMBER"),
			Value:   0,
		},
		&urf.IntFlag{
			Name:    flagIncrement,
			Aliases: []string{"i"},
			Usage:   "line number increment at each line",
			Sources: urf.EnvVars("YUP_NL_LINE_INCREMENT"),
			Value:   0,
		},
		&urf.IntFlag{
			Name:    flagWidth,
			Aliases: []string{"w"},
			Usage:   "use NUMBER columns for line numbers",
			Sources: urf.EnvVars("YUP_NL_NUMBER_WIDTH"),
			Value:   0,
		},
		&urf.StringFlag{
			Name:    flagFormat,
			Aliases: []string{"n"},
			Usage:   "insert line numbers according to FORMAT (ln, rn, rz)",
			Sources: urf.EnvVars("YUP_NL_NUMBER_FORMAT"),
			Value:   "",
		},
	}
}

// build maps the invocation to nl's pipeline: a file-or-stdin source into the nl
// command configured by the numbering flags. An invalid -b or -n is a usage
// error.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	opts, err := options(inv.Args)
	if err != nil {
		return nil, nil, err
	}
	return clix.OperandsOrStdin(inv), command.Nl(opts...), nil
}

func options(c *urf.Command) ([]any, error) {
	opts, err := valueOptions(c)
	if err != nil {
		return nil, err
	}
	opts = append(opts, scalarOptions(c)...)
	return opts, nil
}

// valueOptions collects the flags whose values must be validated against a
// fixed set (-b style, -n format) before mapping to library options.
func valueOptions(c *urf.Command) ([]any, error) {
	var opts []any
	if c.IsSet(flagBody) {
		opt, ok := bodyOptions[c.String(flagBody)]
		if !ok {
			return nil, ErrInvalidBody
		}
		opts = append(opts, opt)
	}
	if c.IsSet(flagFormat) {
		format := c.String(flagFormat)
		if _, ok := formatValues[format]; !ok {
			return nil, ErrInvalidFormat
		}
		opts = append(opts, command.NlFormat(format))
	}
	return opts, nil
}

// scalarOptions collects the flags that map directly to a library option with
// no validation (-s, -v, -i, -w).
func scalarOptions(c *urf.Command) []any {
	var opts []any
	if c.IsSet(flagSep) {
		opts = append(opts, command.NlSep(c.String(flagSep)))
	}
	if c.IsSet(flagStart) {
		opts = append(opts, command.NlStart(c.Int(flagStart)))
	}
	if c.IsSet(flagIncrement) {
		opts = append(opts, command.NlIncrement(c.Int(flagIncrement)))
	}
	if c.IsSet(flagWidth) {
		opts = append(opts, command.NlWidth(c.Int(flagWidth)))
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
