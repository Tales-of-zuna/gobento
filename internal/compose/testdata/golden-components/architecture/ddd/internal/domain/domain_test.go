package domain

import (
	"errors"
	"testing"
)

func TestValidateName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		valid bool
	}{
		{"first", true},
		{"two words", true},
		{"", false},
		{"   ", false},
		{" first", false},
		{"first\n", false},
	}
	for _, tt := range tests {
		err := ValidateName(tt.name)
		if tt.valid && err != nil {
			t.Errorf("ValidateName(%q) error = %v, want nil", tt.name, err)
		}
		if !tt.valid && !errors.Is(err, ErrInvalidName) {
			t.Errorf("ValidateName(%q) error = %v, want ErrInvalidName", tt.name, err)
		}
	}
}
