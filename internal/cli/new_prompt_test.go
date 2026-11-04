package cli

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tales-of-zuna/gobento/components"
	"io/fs"
	"slices"
	"strings"
	"testing"
)

// runInteractive executes gobento with args as in a terminal, answering
// questions with ask.
func runInteractive(t *testing.T, catalog fs.FS, gen *fakeGenerator, ask asker, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	a := &app{
		catalog: catalog, generate: gen.generate, stdout: &stdout, stderr: &stderr,
		interactive: func() bool { return true }, asker: ask,
	}
	cmd := newRootCommand("1.2.3", a)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

// TestNewAsks covers the spec scenarios "Several frameworks, framework flag
// omitted", "No name given", and "Framework chosen by a question".
func TestNewAsks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		answers    map[string]string
		wantAsked  []string
		wantOut    string
		wantStderr string
	}{
		{
			name:       "framework only",
			args:       []string{"new", "shop", "--dry-run"},
			answers:    map[string]string{"framework": "chi"},
			wantAsked:  []string{"framework"},
			wantOut:    "gobento would create shop (mvc/chi/sqlite/sqlc, Go 1.27) in shop",
			wantStderr: "Equivalent command: gobento new shop --framework chi --dry-run\n",
		},
		{
			name:       "name, module, and framework",
			args:       []string{"new", "--dry-run"},
			answers:    map[string]string{"name": "shop", "module": "example.com/shop", "framework": "nethttp"},
			wantAsked:  []string{"name", "module", "framework"},
			wantOut:    "gobento would create example.com/shop (mvc/nethttp/sqlite/sqlc, Go 1.27) in shop",
			wantStderr: "Equivalent command: gobento new shop --module example.com/shop --framework nethttp --dry-run\n",
		},
		{
			name:      "flags skip their questions",
			args:      []string{"new", "shop", "--framework", "nethttp", "--dry-run"},
			wantAsked: nil,
			wantOut:   "(mvc/nethttp/sqlite/sqlc, Go 1.27)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ask := &scriptedAsker{answers: tt.answers}
			gen := &fakeGenerator{}
			out, stderr, err := runInteractive(t, components.FS(), gen, ask, tt.args...)
			if err != nil {
				t.Fatalf("gobento %v error = %v", tt.args, err)
			}
			if !slices.Equal(ask.fields(), tt.wantAsked) {
				t.Errorf("asked %v, want %v", ask.fields(), tt.wantAsked)
			}
			if !strings.Contains(out, tt.wantOut) {
				t.Errorf("output does not contain %q:\n%s", tt.wantOut, out)
			}
			if stderr != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", stderr, tt.wantStderr)
			}
			if len(gen.calls) != 0 {
				t.Errorf("a dry run generated a project")
			}
		})
	}
}

func TestNewAsksThenGenerates(t *testing.T) {
	t.Parallel()

	ask := &scriptedAsker{answers: map[string]string{"name": "shop", "module": "", "framework": "chi"}}
	gen := &fakeGenerator{}
	out, _, err := runInteractive(t, components.FS(), gen, ask, "new")
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.calls) != 1 || gen.calls[0].Target != "shop" || gen.calls[0].Module != "shop" {
		t.Fatalf("generate calls = %+v, want one for module shop in shop", gen.calls)
	}
	if !strings.Contains(out, "Created shop in shop.") {
		t.Errorf("output = %q", out)
	}
}

// TestNewDoesNotAsk covers the spec scenarios "Invalid flag value", "Piped
// input", and "No input flag in a terminal".
func TestNewDoesNotAsk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		interactive bool
		args        []string
		want        string
	}{
		{name: "no input flag", interactive: true, args: []string{"new", "shop", "--no-input"}, want: "--framework: choose a framework; supported: chi, gin, nethttp"},
		{name: "no input flag without a name", interactive: true, args: []string{"new", "--no-input"}, want: "expected exactly one project name, for example: gobento new shop (got 0 arguments)"},
		{name: "no terminal", args: []string{"new", "shop"}, want: "--framework: choose a framework; supported: chi, gin, nethttp"},
		{name: "no terminal without a name", args: []string{"new"}, want: "expected exactly one project name, for example: gobento new shop (got 0 arguments)"},
		{name: "invalid flag", interactive: true, args: []string{"new", "shop", "--framework", "express"}, want: `--framework: unknown framework "express"; supported: chi, gin, nethttp`},
		{name: "invalid name", interactive: true, args: []string{"new", "My App"}, want: `invalid project name "My App"`},
		{name: "invalid pin", interactive: true, args: []string{"new", "shop", "--pin", "nonsense"}, want: `--pin "nonsense": want module=version`},
		{name: "invalid flag without a name", interactive: true, args: []string{"new", "--framework", "express"}, want: "expected exactly one project name"},
		{name: "two names", interactive: true, args: []string{"new", "a", "b"}, want: "expected exactly one project name, for example: gobento new shop (got 2 arguments)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ask := &scriptedAsker{}
			gen := &fakeGenerator{}
			var stdout, stderr bytes.Buffer
			a := &app{
				catalog: components.FS(), generate: gen.generate, stdout: &stdout, stderr: &stderr,
				interactive: func() bool { return tt.interactive }, asker: ask,
			}
			cmd := newRootCommand("dev", a)
			cmd.SetArgs(tt.args)
			err := cmd.ExecuteContext(context.Background())
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want it to contain %q", err, tt.want)
			}
			if len(ask.asked) != 0 || len(gen.calls) != 0 || stderr.Len() != 0 {
				t.Errorf("asked %v, generated %d times, stderr %q; want nothing", ask.fields(), len(gen.calls), stderr.String())
			}
		})
	}
}

// TestNewAborted covers the spec scenarios "Ctrl-C at a question" and
// "Input ends".
func TestNewAborted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err        error
		wantStatus int
		wantStderr string
	}{
		{err: errAborted, wantStatus: ExitInterrupted, wantStderr: "gobento: interrupted; nothing was left behind\n"},
		{err: errInputEnded, wantStatus: 1, wantStderr: "gobento: input ended before every question was answered; nothing was written\n"},
	}
	for _, tt := range tests {
		ask := &scriptedAsker{errs: map[string]error{"framework": tt.err}}
		gen := &fakeGenerator{}
		var stdout, stderr bytes.Buffer
		a := &app{
			catalog: components.FS(), generate: gen.generate, stdout: &stdout, stderr: &stderr,
			interactive: func() bool { return true }, asker: ask,
		}
		cmd := newRootCommand("dev", a)
		cmd.SetArgs([]string{"new", "shop"})
		ctx := context.Background()
		err := cmd.ExecuteContext(ctx)
		if !errors.Is(err, tt.err) {
			t.Errorf("error = %v, want %v", err, tt.err)
		}
		if got := a.exitStatus(ctx, err); got != tt.wantStatus {
			t.Errorf("%v: exit status = %d, want %d", tt.err, got, tt.wantStatus)
		}
		if stderr.String() != tt.wantStderr || stdout.Len() != 0 || len(gen.calls) != 0 {
			t.Errorf("%v: stderr %q, stdout %q, %d generations; want only %q", tt.err, stderr.String(), stdout.String(), len(gen.calls), tt.wantStderr)
		}
	}
}
