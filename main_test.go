package main

import (
	"context"
	"errors"
	"testing"

	clix "github.com/gloo-foo/cli"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// parse runs args through a bare command carrying a fresh set of the wrapper's
// flags and returns the parsed accessor.
func parse(t *testing.T, args ...string) *urf.Command {
	t.Helper()
	var got *urf.Command
	app := &urf.Command{
		Name:   name,
		Flags:  flags(),
		Action: func(_ context.Context, c *urf.Command) error { got = c; return nil },
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func TestOptions(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"none", []string{name}, 0},
		{"body", []string{name, "-b", "t"}, 1},
		{"format", []string{name, "-n", "rz"}, 1},
		{"sep", []string{name, "-s", ": "}, 1},
		{"start", []string{name, "-v", "3"}, 1},
		{"increment", []string{name, "-i", "2"}, 1},
		{"width", []string{name, "-w", "6"}, 1},
		{"all", []string{name, "-b", "a", "-n", "ln", "-s", "|", "-v", "1", "-i", "1", "-w", "4"}, 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts, err := options(parse(t, tc.args...))
			if err != nil {
				t.Fatalf("options err=%v", err)
			}
			if len(opts) != tc.want {
				t.Fatalf("options len=%d, want %d", len(opts), tc.want)
			}
		})
	}
}

func TestOptionsErrors(t *testing.T) {
	if _, err := options(parse(t, name, "-b", "x")); !errors.Is(err, ErrInvalidBody) {
		t.Fatalf("err=%v, want ErrInvalidBody", err)
	}
	if _, err := options(parse(t, name, "-n", "zz")); !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("err=%v, want ErrInvalidFormat", err)
	}
}

func TestErrorMessages(t *testing.T) {
	if ErrInvalidBody.Error() != string(ErrInvalidBody) {
		t.Fatalf("ErrInvalidBody message=%q", ErrInvalidBody.Error())
	}
	if ErrInvalidFormat.Error() != string(ErrInvalidFormat) {
		t.Fatalf("ErrInvalidFormat message=%q", ErrInvalidFormat.Error())
	}
}

func TestBuild(t *testing.T) {
	src, filter, err := build(clix.Invocation{Args: parse(t, name), Fs: afero.NewMemMapFs()})
	if err != nil || src == nil || filter == nil {
		t.Fatalf("build: src=%v filter=%v err=%v", src, filter, err)
	}
}

func TestBuild_InvalidBodyIsUsageError(t *testing.T) {
	src, filter, err := build(clix.Invocation{Args: parse(t, name, "-b", "x"), Fs: afero.NewMemMapFs()})
	if !errors.Is(err, ErrInvalidBody) {
		t.Fatalf("err=%v, want ErrInvalidBody", err)
	}
	if src != nil || filter != nil {
		t.Fatalf("src=%v filter=%v, want both nil on error", src, filter)
	}
}

func Test_main(t *testing.T) {
	orig := runMain
	t.Cleanup(func() { runMain = orig })
	var gotName clix.Name
	runMain = func(s clix.Spec, _ clix.Version) { gotName = s.Name }
	main()
	if gotName != name {
		t.Fatalf("main used spec %q, want %s", gotName, name)
	}
}
