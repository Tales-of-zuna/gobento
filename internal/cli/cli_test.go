package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandVersion(t *testing.T) {
	t.Parallel()

	cmd := NewRootCommand("1.2.3")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, "1.2.3") {
		t.Errorf("--version output = %q, want it to contain %q", got, "1.2.3")
	}
}
