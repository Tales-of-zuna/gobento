package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"github.com/charmbracelet/colorprofile"
)

// huhAsker asks questions with huh: a full-screen form in a terminal, or
// plain line prompts in accessible mode.
type huhAsker struct {
	in         io.Reader
	out        io.Writer
	accessible bool
	// environ is the environment that decides how accessible output is
	// styled; nil means the process environment.
	environ []string
	// width is the width of the full-screen form; 0 means the terminal's.
	width int
}

// ask asks q with a form of one field.
func (h huhAsker) ask(ctx context.Context, q question) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var value string
	var field huh.Field
	if len(q.options) > 0 {
		opts := make([]huh.Option[string], len(q.options))
		for i, o := range q.options {
			opts[i] = huh.NewOption(o.label, o.value)
		}
		field = huh.NewSelect[string]().Title(q.title).Description(q.help).Options(opts...).Value(&value)
	} else {
		validate := func(s string) error {
			if s == "" {
				s = q.value
			}
			if q.validate == nil {
				return nil
			}
			return q.validate(s)
		}
		// The default is a placeholder, not editable text, so that typing
		// replaces it and Enter keeps it.
		field = huh.NewInput().Title(q.title).Description(q.help).Placeholder(q.value).Value(&value).Validate(validate)
	}
	form := huh.NewForm(huh.NewGroup(field)).WithKeyMap(keyMap()).WithAccessible(h.accessible)

	var err error
	if h.accessible {
		err = h.runAccessible(ctx, form, q.help)
	} else {
		if h.width > 0 {
			form = form.WithWidth(h.width)
		}
		err = form.WithInput(h.in).WithOutput(h.out).RunWithContext(ctx)
		switch {
		case ctx.Err() != nil:
			err = ctx.Err()
		case errors.Is(err, huh.ErrUserAborted):
			err = errAborted
		}
	}
	if err != nil {
		return "", err
	}
	if value == "" {
		value = q.value
	}
	return value, nil
}

// runAccessible runs form in accessible mode until it ends or ctx is done.
// huh's line prompts cannot be cancelled and treat the end of input as an
// empty answer, so the form runs in a goroutine and reads through an
// eofReader.
func (h huhAsker) runAccessible(ctx context.Context, form *huh.Form, help string) error {
	environ := h.environ
	if environ == nil {
		environ = os.Environ()
	}
	in := &eofReader{r: h.in}
	out := colorprofile.NewWriter(h.out, environ)
	form = form.WithInput(in).WithOutput(out)
	// Accessible fields print only their title, so the hint comes first.
	if help != "" {
		_, _ = fmt.Fprintln(out, help)
	}
	done := make(chan error, 1)
	go func() { done <- form.Run() }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err == nil && in.ended() {
			err = errInputEnded
		}
		return err
	}
}

// keyMap returns huh's default keys with Esc added to Ctrl-C for quitting.
// The option filter, which also uses Esc, is turned off; the lists are
// short.
func keyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("esc", "quit"))
	km.Select.Filter.SetEnabled(false)
	return km
}

// eofReader reads one byte at a time and records the end of input. huh
// reads each accessible answer with a new bufio.Scanner, which would
// otherwise buffer, and lose, the answers that follow in the same read.
type eofReader struct {
	r   io.Reader
	mu  sync.Mutex
	eof bool
}

func (e *eofReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	n, err := e.r.Read(p)
	if err != nil {
		e.mu.Lock()
		e.eof = true
		e.mu.Unlock()
	}
	return n, err //nolint:wrapcheck // a reader passes its source's errors on
}

// ended reports whether the input ended or failed.
func (e *eofReader) ended() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.eof
}
