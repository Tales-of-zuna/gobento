package compat

import (
	"errors"
	"testing"
)

func TestNormalizeGo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "1.27", want: "v1.27.0"},
		{in: "1.26.0", want: "v1.26.0"},
		{in: "1.27.3", want: "v1.27.3"},
		{in: "1.100", want: "v1.100.0"},
		{in: "", wantErr: true},
		{in: "1", wantErr: true},
		{in: "v1.27", wantErr: true},
		{in: "go1.27", wantErr: true},
		{in: "1.27rc1", wantErr: true},
		{in: "1.27.0.1", wantErr: true},
		{in: "1.027", wantErr: true},
		{in: "1..2", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeGo(tt.in)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidVersion) {
					t.Fatalf("NormalizeGo(%q) error = %v, want ErrInvalidVersion", tt.in, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeGo(%q) error = %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("NormalizeGo(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCheckModuleVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      string
		wantErr bool
	}{
		{in: "v1.61.1"},
		{in: "v2.0.0-rc.1"},
		{in: "v0.0.0-20260101000000-abcdefabcdef"},
		{in: "1.61.1", wantErr: true},
		{in: "v1.61", wantErr: true},
		{in: "v1", wantErr: true},
		{in: "v1.61.1+meta", wantErr: true},
		{in: "latest", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			err := CheckModuleVersion(tt.in)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("CheckModuleVersion(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrInvalidVersion) {
				t.Errorf("CheckModuleVersion(%q) error = %v, want ErrInvalidVersion", tt.in, err)
			}
		})
	}
}

func TestRangeContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rng     string
		version string
		want    bool
	}{
		{name: ">= equal", rng: ">=v1.2.0", version: "v1.2.0", want: true},
		{name: ">= above", rng: ">=v1.2.0", version: "v1.3.0", want: true},
		{name: ">= below", rng: ">=v1.2.0", version: "v1.1.9", want: false},
		{name: "> equal", rng: ">v1.2.0", version: "v1.2.0", want: false},
		{name: "> above", rng: ">v1.2.0", version: "v1.2.1", want: true},
		{name: "<= equal", rng: "<=v1.2.0", version: "v1.2.0", want: true},
		{name: "<= above", rng: "<=v1.2.0", version: "v1.2.1", want: false},
		{name: "< equal", rng: "<v2.0.0", version: "v2.0.0", want: false},
		{name: "< below", rng: "<v2.0.0", version: "v1.99.0", want: true},
		{name: "< prerelease of bound", rng: "<v2.0.0", version: "v2.0.0-rc.1", want: true},
		{name: "= equal", rng: "=v1.31.1", version: "v1.31.1", want: true},
		{name: "= different", rng: "=v1.31.1", version: "v1.31.0", want: false},
		{name: "and inside", rng: ">=v1.61.0 <v2.0.0", version: "v1.61.1", want: true},
		{name: "and above upper", rng: ">=v1.61.0 <v2.0.0", version: "v2.0.0", want: false},
		{name: "and below lower", rng: ">=v1.61.0 <v2.0.0", version: "v1.60.1", want: false},
		{name: "extra spaces", rng: "  >=v1.0.0   <v1.1.0 ", version: "v1.0.5", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := ParseRange(tt.rng)
			if err != nil {
				t.Fatalf("ParseRange(%q) error = %v", tt.rng, err)
			}
			if got := r.Contains(tt.version); got != tt.want {
				t.Errorf("Range(%q).Contains(%q) = %v, want %v", tt.rng, tt.version, got, tt.want)
			}
		})
	}
}

func TestParseRangeMalformed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rng  string
	}{
		{name: "empty", rng: ""},
		{name: "only spaces", rng: "   "},
		{name: "missing operator", rng: "v1.2.0"},
		{name: "unknown operator", rng: "~v1.2.0"},
		{name: "double equals", rng: "==v1.2.0"},
		{name: "missing version", rng: ">="},
		{name: "space after operator", rng: ">= v1.2.0"},
		{name: "missing v prefix", rng: ">=1.2.0"},
		{name: "shorthand version", rng: ">=v1.2"},
		{name: "one bad comparator", rng: ">=v1.2.0 <vX"},
		{name: "comma separated", rng: ">=v1.2.0,<v2.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := ParseRange(tt.rng); !errors.Is(err, ErrInvalidRange) {
				t.Errorf("ParseRange(%q) error = %v, want ErrInvalidRange", tt.rng, err)
			}
		})
	}
}

func TestParseGoRange(t *testing.T) {
	t.Parallel()

	r, err := ParseGoRange(">=1.26 <1.28")
	if err != nil {
		t.Fatalf("ParseGoRange() error = %v", err)
	}
	tests := []struct {
		goVersion string
		want      bool
	}{
		{goVersion: "1.25.9", want: false},
		{goVersion: "1.26", want: true},
		{goVersion: "1.27.3", want: true},
		{goVersion: "1.28", want: false},
	}
	for _, tt := range tests {
		v, err := NormalizeGo(tt.goVersion)
		if err != nil {
			t.Fatalf("NormalizeGo(%q) error = %v", tt.goVersion, err)
		}
		if got := r.Contains(v); got != tt.want {
			t.Errorf("Contains(%s) = %v, want %v", tt.goVersion, got, tt.want)
		}
	}

	for _, bad := range []string{">=v1.26", ">=1.26rc1", "1.26"} {
		if _, err := ParseGoRange(bad); !errors.Is(err, ErrInvalidRange) {
			t.Errorf("ParseGoRange(%q) error = %v, want ErrInvalidRange", bad, err)
		}
	}
}

func TestRangeExactAndString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rng       string
		wantExact string
		wantOK    bool
	}{
		{rng: "=v1.31.1", wantExact: "v1.31.1", wantOK: true},
		{rng: ">=v1.31.1", wantOK: false},
		{rng: "=v1.31.1 <v2.0.0", wantOK: false},
	}
	for _, tt := range tests {
		r, err := ParseRange(tt.rng)
		if err != nil {
			t.Fatalf("ParseRange(%q) error = %v", tt.rng, err)
		}
		got, ok := r.Exact()
		if got != tt.wantExact || ok != tt.wantOK {
			t.Errorf("Range(%q).Exact() = %q, %v, want %q, %v", tt.rng, got, ok, tt.wantExact, tt.wantOK)
		}
		if r.String() != tt.rng {
			t.Errorf("Range(%q).String() = %q", tt.rng, r.String())
		}
	}
}

func TestGoDirective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{in: "1.27", want: "1.27.0"},
		{in: "1.26", want: "1.26.0"},
		{in: "1.26.3", want: "1.26.3"},
		{in: "go1.27", wantErr: true},
	}
	for _, tt := range tests {
		got, err := GoDirective(tt.in)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("GoDirective(%q) = %q, %v; want %q, wantErr %v", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}

func TestRangeFloor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rng    string
		want   string
		wantOK bool
	}{
		{rng: ">=v1.61.0 <v2.0.0", want: "v1.61.0", wantOK: true},
		{rng: "=v1.31.1", want: "v1.31.1", wantOK: true},
		{rng: ">=v1.0.0 >=v1.2.0", want: "v1.2.0", wantOK: true},
		{rng: ">v1.0.0", wantOK: false},
		{rng: "<v2.0.0", wantOK: false},
		{rng: ">=v2.0.0 <v1.0.0", wantOK: false},
	}
	for _, tt := range tests {
		r, err := ParseRange(tt.rng)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := r.Floor()
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("Range(%q).Floor() = %q, %v; want %q, %v", tt.rng, got, ok, tt.want, tt.wantOK)
		}
	}
	g, _ := ParseGoRange(">=1.26 <1.28")
	if got, ok := g.Floor(); got != "v1.26.0" || !ok {
		t.Errorf("go range Floor() = %q, %v; want v1.26.0, true", got, ok)
	}
}
