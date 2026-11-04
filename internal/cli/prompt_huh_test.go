package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// accessibleAsker returns a huhAsker in accessible mode that reads input
// and writes to out without colors.
func accessibleAsker(input io.Reader, out io.Writer) huhAsker {
	return huhAsker{in: input, out: out, accessible: true, environ: []string{"TERM=dumb"}}
}

var frameworkQuestion = kindQuestion("framework", []option{
	{label: "chi", value: "chi"},
	{label: "gin", value: "gin"},
	{label: "net/http (nethttp)", value: "nethttp"},
})

func TestHuhAskerAccessibleChoice(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	got, err := accessibleAsker(strings.NewReader("3\n"), &out).ask(context.Background(), frameworkQuestion)
	if err != nil {
		t.Fatalf("ask() error = %v", err)
	}
	if got != "nethttp" {
		t.Errorf("ask() = %q, want nethttp", got)
	}
	for _, want := range []string{"HTTP framework", "1. chi", "2. gin", "3. net/http (nethttp)"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output does not contain %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Errorf("accessible output contains escape sequences:\n%q", out.String())
	}
}

// TestHuhAskerAccessibleSequence answers several questions from one input,
// as a pasted or fast-typed answer arrives, and re-asks after an invalid
// name.
func TestHuhAskerAccessibleSequence(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	a := accessibleAsker(strings.NewReader("My App\nshop\n\n2\n"), &out)
	name, err := a.ask(context.Background(), nameQuestion())
	if err != nil || name != "shop" {
		t.Fatalf("name = %q, %v; want shop", name, err)
	}
	if !strings.Contains(out.String(), `invalid project name "My App"`) {
		t.Errorf("no explanation for the invalid name:\n%s", out.String())
	}
	module, err := a.ask(context.Background(), moduleQuestion(name))
	if err != nil || module != "shop" {
		t.Errorf("module = %q, %v; want the name for an empty answer", module, err)
	}
	if !strings.Contains(out.String(), "press Enter to use shop") {
		t.Errorf("the module question's hint is missing:\n%s", out.String())
	}
	framework, err := a.ask(context.Background(), frameworkQuestion)
	if err != nil || framework != "gin" {
		t.Errorf("framework = %q, %v; want gin", framework, err)
	}
}

func TestHuhAskerAccessibleInputEnds(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"", "My App\n"} {
		_, err := accessibleAsker(strings.NewReader(input), io.Discard).ask(context.Background(), nameQuestion())
		if !errors.Is(err, errInputEnded) {
			t.Errorf("input %q: ask() error = %v, want errInputEnded", input, err)
		}
	}
	_, err := accessibleAsker(strings.NewReader(""), io.Discard).ask(context.Background(), frameworkQuestion)
	if !errors.Is(err, errInputEnded) {
		t.Errorf("choice at the end of input: ask() error = %v, want errInputEnded", err)
	}
}

// TestHuhAskerAccessibleCancel covers Ctrl-C in accessible mode, where the
// signal cancels the context while a line is being read.
func TestHuhAskerAccessibleCancel(t *testing.T) {
	t.Parallel()

	r, w := io.Pipe()
	defer w.Close() //nolint:errcheck // test pipe
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := accessibleAsker(r, io.Discard).ask(ctx, frameworkQuestion)
		done <- err
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("ask() error = %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ask() did not return after the context was cancelled")
	}
}

func TestHuhAskerCancelledBefore(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, accessible := range []bool{false, true} {
		a := huhAsker{in: strings.NewReader("1\n"), out: io.Discard, accessible: accessible}
		if _, err := a.ask(ctx, frameworkQuestion); !errors.Is(err, context.Canceled) {
			t.Errorf("accessible %v: ask() error = %v, want context.Canceled", accessible, err)
		}
	}
}

func TestKeyMapQuitsOnEsc(t *testing.T) {
	t.Parallel()

	km := keyMap()
	if keys := km.Quit.Keys(); !strings.Contains(strings.Join(keys, " "), "esc") || !strings.Contains(strings.Join(keys, " "), "ctrl+c") {
		t.Errorf("quit keys = %v, want ctrl+c and esc", keys)
	}
	if km.Select.Filter.Enabled() {
		t.Error("the select filter is enabled; its Esc would not quit")
	}
}

func TestAccessible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env  map[string]string
		want bool
	}{
		{env: map[string]string{}, want: false},
		{env: map[string]string{"TERM": "xterm-256color"}, want: false},
		{env: map[string]string{"ACCESSIBLE": "1"}, want: true},
		{env: map[string]string{"TERM": "dumb"}, want: true},
	}
	for _, tt := range tests {
		a := &app{getenv: func(k string) string { return tt.env[k] }}
		if got := a.accessible(); got != tt.want {
			t.Errorf("env %v: accessible() = %v, want %v", tt.env, got, tt.want)
		}
	}
	if (&app{}).accessible() {
		t.Error("an app without an environment is accessible")
	}
}

// TestHuhAskerFullScreen drives the full-screen form with key presses, as a
// terminal in raw mode sends them. The width stands in for the terminal
// size that a test's input cannot report.
func TestHuhAskerFullScreen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		q       question
		keys    string
		want    string
		wantErr error
	}{
		{name: "choice", q: frameworkQuestion, keys: "j\r", want: "gin"},
		{name: "first choice", q: frameworkQuestion, keys: "\r", want: "chi"},
		{name: "esc", q: frameworkQuestion, keys: "\x1b", wantErr: errAborted},
		{name: "ctrl-c", q: frameworkQuestion, keys: "\x03", wantErr: errAborted},
		{name: "typed text", q: moduleQuestion("shop"), keys: "example.com/shop\r", want: "example.com/shop"},
		{name: "default text", q: moduleQuestion("shop"), keys: "\r", want: "shop"},
		{name: "ctrl-c in text", q: nameQuestion(), keys: "sh\x03", wantErr: errAborted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			got, err := huhAsker{in: strings.NewReader(tt.keys), out: io.Discard, width: 80}.ask(ctx, tt.q)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ask() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("ask() = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

func TestFormWidth(t *testing.T) {
	t.Parallel()

	if got := formWidth(io.Discard); got != 0 {
		t.Errorf("formWidth(non-file) = %d, want 0", got)
	}
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() //nolint:errcheck // test file
	if got := formWidth(f); got != 80 {
		t.Errorf("formWidth(file without a size) = %d, want 80", got)
	}
}
