package cli

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// scriptedAsker answers questions from a script keyed by question field and
// records every question it was asked.
type scriptedAsker struct {
	answers map[string]string
	errs    map[string]error
	asked   []question
}

func (s *scriptedAsker) ask(_ context.Context, q question) (string, error) {
	s.asked = append(s.asked, q)
	if err := s.errs[q.field]; err != nil {
		return "", err
	}
	answer, ok := s.answers[q.field]
	if !ok {
		return "", fmt.Errorf("unexpected question %q", q.field)
	}
	if len(q.options) > 0 && !slices.ContainsFunc(q.options, func(o option) bool { return o.value == answer }) {
		return "", fmt.Errorf("question %q does not offer %q", q.field, answer)
	}
	if answer == "" {
		answer = q.value
	}
	if q.validate != nil {
		if err := q.validate(answer); err != nil {
			return "", err
		}
	}
	return answer, nil
}

// fields returns the fields of the asked questions, in order.
func (s *scriptedAsker) fields() []string {
	var out []string
	for _, q := range s.asked {
		out = append(out, q.field)
	}
	return out
}

// options returns the option values of the question for field.
func (s *scriptedAsker) options(field string) []string {
	for _, q := range s.asked {
		if q.field == field {
			var out []string
			for _, o := range q.options {
				out = append(out, o.value)
			}
			return out
		}
	}
	return nil
}

// selections builds supported selections from "arch/framework/db/data"
// strings, with an optional "/frontend", sharing one component per kind and
// name.
func selections(specs ...string) []catalog.Selection {
	comps := map[string]*catalog.Component{}
	get := func(kind catalog.Kind, name string) *catalog.Component {
		key := string(kind) + "/" + name
		if comps[key] == nil {
			comps[key] = &catalog.Component{Kind: kind, Name: name, Title: strings.ToUpper(name)}
		}
		return comps[key]
	}
	var out []catalog.Selection
	for _, spec := range specs {
		n := strings.Split(spec, "/")
		sel := catalog.Selection{
			Architecture: get(catalog.KindArchitecture, n[0]),
			Framework:    get(catalog.KindFramework, n[1]),
			Database:     get(catalog.KindDatabase, n[2]),
			Data:         get(catalog.KindData, n[3]),
		}
		if len(n) > 4 {
			sel.Frontend = get(catalog.KindFrontend, n[4])
		}
		out = append(out, sel)
	}
	return out
}

// testSupported has three frameworks; only gin also works with postgres
// and gorm.
var testSupported = selections(
	"mvc/chi/sqlite/sqlc",
	"mvc/gin/sqlite/sqlc",
	"mvc/gin/postgres/gorm",
	"mvc/gin/postgres/sqlc",
	"mvc/nethttp/sqlite/sqlc",
)

func TestInterview(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		in          newInput
		answers     map[string]string
		wantAsked   []string
		wantOptions map[string][]string
		want        map[catalog.Kind]string
		wantName    string
		wantModule  string
	}{
		{
			name:        "kinds with several options only",
			in:          newInput{name: "shop", hasName: true},
			answers:     map[string]string{"framework": "nethttp"},
			wantAsked:   []string{"framework"},
			wantOptions: map[string][]string{"framework": {"chi", "gin", "nethttp"}},
			want:        map[catalog.Kind]string{"architecture": "mvc", "framework": "nethttp", "database": "sqlite", "data": "sqlc"},
			wantName:    "shop",
		},
		{
			name:        "later options follow earlier answers",
			in:          newInput{name: "shop", hasName: true},
			answers:     map[string]string{"framework": "gin", "database": "postgres", "data": "gorm"},
			wantAsked:   []string{"framework", "database", "data"},
			wantOptions: map[string][]string{"database": {"postgres", "sqlite"}, "data": {"gorm", "sqlc"}},
			want:        map[catalog.Kind]string{"architecture": "mvc", "framework": "gin", "database": "postgres", "data": "gorm"},
			wantName:    "shop",
		},
		{
			name:      "a flag skips its question and narrows the rest",
			in:        newInput{name: "shop", hasName: true, components: map[catalog.Kind]string{"database": "postgres"}},
			answers:   map[string]string{"data": "sqlc"},
			wantAsked: []string{"data"},
			want:      map[catalog.Kind]string{"architecture": "mvc", "framework": "gin", "database": "postgres", "data": "sqlc"},
			wantName:  "shop",
		},
		{
			name:       "name and then module",
			in:         newInput{components: map[catalog.Kind]string{"framework": "chi"}},
			answers:    map[string]string{"name": "shop", "module": "example.com/shop"},
			wantAsked:  []string{"name", "module"},
			want:       map[catalog.Kind]string{"architecture": "mvc", "framework": "chi", "database": "sqlite", "data": "sqlc"},
			wantName:   "shop",
			wantModule: "example.com/shop",
		},
		{
			name:      "empty module keeps the name",
			in:        newInput{components: map[catalog.Kind]string{"framework": "chi"}},
			answers:   map[string]string{"name": "shop", "module": ""},
			wantAsked: []string{"name", "module"},
			want:      map[catalog.Kind]string{"architecture": "mvc", "framework": "chi", "database": "sqlite", "data": "sqlc"},
			wantName:  "shop",
		},
		{
			name:       "a given module is not asked",
			in:         newInput{module: "example.com/x", hasModule: true, components: map[catalog.Kind]string{"framework": "chi"}},
			answers:    map[string]string{"name": "shop"},
			wantAsked:  []string{"name"},
			want:       map[catalog.Kind]string{"architecture": "mvc", "framework": "chi", "database": "sqlite", "data": "sqlc"},
			wantName:   "shop",
			wantModule: "example.com/x",
		},
		{
			name:     "nothing to ask",
			in:       newInput{name: "shop", hasName: true, components: map[catalog.Kind]string{"framework": "gin", "database": "sqlite"}},
			want:     map[catalog.Kind]string{"architecture": "mvc", "framework": "gin", "database": "sqlite", "data": "sqlc"},
			wantName: "shop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := &scriptedAsker{answers: tt.answers}
			in := tt.in
			asked, err := interview(context.Background(), testSupported, &in, a)
			if err != nil {
				t.Fatalf("interview() error = %v", err)
			}
			if !slices.Equal(a.fields(), tt.wantAsked) {
				t.Errorf("asked %v, want %v", a.fields(), tt.wantAsked)
			}
			if asked != (len(tt.wantAsked) > 0) {
				t.Errorf("interview() asked = %v, want %v", asked, len(tt.wantAsked) > 0)
			}
			for field, want := range tt.wantOptions {
				if got := a.options(field); !slices.Equal(got, want) {
					t.Errorf("%s options = %v, want %v", field, got, want)
				}
			}
			if !maps.Equal(in.components, tt.want) {
				t.Errorf("components = %v, want %v", in.components, tt.want)
			}
			if in.name != tt.wantName || in.module != tt.wantModule || !in.hasName {
				t.Errorf("name %q module %q, want %q %q", in.name, in.module, tt.wantName, tt.wantModule)
			}
		})
	}
}

func TestInterviewQuestions(t *testing.T) {
	t.Parallel()

	a := &scriptedAsker{answers: map[string]string{"name": "shop", "module": "", "framework": "chi"}}
	in := newInput{}
	if _, err := interview(context.Background(), testSupported, &in, a); err != nil {
		t.Fatal(err)
	}
	name, module, framework := a.asked[0], a.asked[1], a.asked[2]
	if name.validate("My App") == nil || name.validate("con") == nil || name.validate("shop") != nil {
		t.Error("the name question does not apply the project name rules")
	}
	if module.value != "shop" || module.validate("Not A Path") == nil || module.validate("example.com/shop") != nil {
		t.Errorf("the module question has default %q or does not apply the module path rules", module.value)
	}
	if framework.title != "HTTP framework" || !strings.Contains(framework.help, "--framework") {
		t.Errorf("framework question = %q / %q, want its title and flag", framework.title, framework.help)
	}
	if got := framework.options[0]; got.label != "CHI (chi)" || got.value != "chi" {
		t.Errorf("first framework option = %+v, want the title and name of chi", got)
	}
}

func TestInterviewStopsOnError(t *testing.T) {
	t.Parallel()

	for _, want := range []error{errAborted, errInputEnded, context.Canceled} {
		a := &scriptedAsker{
			answers: map[string]string{"name": "shop", "module": ""},
			errs:    map[string]error{"framework": want},
		}
		in := newInput{}
		if _, err := interview(context.Background(), testSupported, &in, a); !errors.Is(err, want) {
			t.Errorf("interview() error = %v, want %v", err, want)
		}
		if got := a.fields(); !slices.Equal(got, []string{"name", "module", "framework"}) {
			t.Errorf("asked %v after %v, want it to stop at the framework", got, want)
		}
	}
}

func TestCanAsk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      newInput
		pinErrs []error
		want    bool
	}{
		{name: "nothing given", in: newInput{}, want: true},
		{name: "valid name", in: newInput{name: "shop", hasName: true}, want: true},
		{name: "invalid name", in: newInput{name: "My App", hasName: true}},
		{name: "invalid module", in: newInput{name: "shop", hasName: true, module: "Not A Path", hasModule: true}},
		{name: "invalid pin", in: newInput{name: "shop", hasName: true}, pinErrs: []error{errors.New("bad pin")}},
		{name: "unknown framework", in: newInput{name: "shop", hasName: true, components: map[catalog.Kind]string{"framework": "express"}}},
		{name: "unsupported combination", in: newInput{name: "shop", hasName: true, components: map[catalog.Kind]string{"framework": "chi", "data": "gorm"}}},
		{name: "supported flags", in: newInput{name: "shop", hasName: true, components: map[catalog.Kind]string{"framework": "gin", "data": "gorm"}}, want: true},
	}
	for _, tt := range tests {
		if got := canAsk(testSupported, tt.in, tt.pinErrs); got != tt.want {
			t.Errorf("%s: canAsk() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// withFrontend is testSupported where every selection also comes with the
// frontend htmx.
var withFrontend = selections(
	"mvc/chi/sqlite/sqlc", "mvc/chi/sqlite/sqlc/htmx",
	"mvc/gin/sqlite/sqlc", "mvc/gin/sqlite/sqlc/htmx",
)

// TestInterviewFrontend covers the spec scenario "Frontend question": the
// optional kind is asked with "None (API only)" first.
func TestInterviewFrontend(t *testing.T) {
	t.Parallel()

	for answer, want := range map[string]map[catalog.Kind]string{
		"none": {"architecture": "mvc", "framework": "gin", "database": "sqlite", "data": "sqlc", "frontend": "none"},
		"htmx": {"architecture": "mvc", "framework": "gin", "database": "sqlite", "data": "sqlc", "frontend": "htmx"},
	} {
		a := &scriptedAsker{answers: map[string]string{"framework": "gin", "frontend": answer}}
		in := newInput{name: "shop", hasName: true}
		if _, err := interview(context.Background(), withFrontend, &in, a); err != nil {
			t.Fatalf("interview() error = %v", err)
		}
		if got := a.options("frontend"); !slices.Equal(got, []string{"none", "htmx"}) {
			t.Errorf("frontend options = %v, want [none htmx]", got)
		}
		if q := a.asked[len(a.asked)-1]; q.options[0].label != "None (API only)" || q.title != "Frontend" {
			t.Errorf("frontend question = %q with first option %q", q.title, q.options[0].label)
		}
		if !maps.Equal(in.components, want) {
			t.Errorf("answer %s: components = %v, want %v", answer, in.components, want)
		}
	}

	// A frontend flag narrows the other kinds like any flag.
	a := &scriptedAsker{answers: map[string]string{"framework": "chi"}}
	in := newInput{name: "shop", hasName: true, components: map[catalog.Kind]string{"frontend": "none"}}
	if _, err := interview(context.Background(), withFrontend, &in, a); err != nil {
		t.Fatalf("interview() error = %v", err)
	}
	if !slices.Equal(a.fields(), []string{"framework"}) {
		t.Errorf("asked %v with --frontend none, want only the framework", a.fields())
	}
}

func TestMatchingNone(t *testing.T) {
	t.Parallel()

	for chosen, want := range map[string]int{"": 4, "none": 2, "htmx": 2, "react": 0} {
		if got := len(matching(withFrontend, map[catalog.Kind]string{"frontend": chosen})); got != want {
			t.Errorf("matching(frontend=%q) = %d selections, want %d", chosen, got, want)
		}
	}
}

// TestEquivalentCommandFrontend covers the spec scenario "No frontend chosen
// by a question".
func TestEquivalentCommandFrontend(t *testing.T) {
	t.Parallel()

	components := func(frontend string) map[catalog.Kind]string {
		return map[catalog.Kind]string{"architecture": "mvc", "framework": "gin", "database": "sqlite", "data": "sqlc", "frontend": frontend}
	}
	tests := []struct {
		name     string
		frontend string
		f        newFlags
		want     string
	}{
		{"none answered", "none", newFlags{}, "gobento new shop --framework gin"},
		{"none given", "none", newFlags{frontend: "none"}, "gobento new shop --framework gin --frontend none"},
		{"htmx answered", "htmx", newFlags{}, "gobento new shop --framework gin --frontend htmx"},
	}
	for _, tt := range tests {
		in := newInput{name: "shop", hasName: true, components: components(tt.frontend)}
		if got := equivalentCommand(withFrontend, in, tt.f); got != tt.want {
			t.Errorf("%s:\ngot  %s\nwant %s", tt.name, got, tt.want)
		}
	}
}

func TestEquivalentCommand(t *testing.T) {
	t.Parallel()

	all := map[catalog.Kind]string{"architecture": "mvc", "framework": "gin", "database": "sqlite", "data": "sqlc"}
	tests := []struct {
		name string
		in   newInput
		f    newFlags
		want string
	}{
		{
			name: "only kinds with several options",
			in:   newInput{name: "shop", hasName: true, components: all},
			want: "gobento new shop --framework gin --db sqlite --data sqlc",
		},
		{
			name: "a given single-option kind and the other flags",
			in:   newInput{name: "shop", hasName: true, module: "example.com/my shop", hasModule: true, components: all},
			f:    newFlags{arch: "mvc", goVersion: "1.26", pins: []string{"modernc.org/sqlite=v1.61.0"}, dir: "out dir", dryRun: true},
			want: "gobento new shop --module 'example.com/my shop' --arch mvc --framework gin --db sqlite --data sqlc " +
				"--go 1.26 --pin modernc.org/sqlite=v1.61.0 --dir 'out dir' --dry-run",
		},
		{
			name: "module equal to the name is left out",
			in:   newInput{name: "shop", hasName: true, module: "shop", hasModule: true, components: all},
			want: "gobento new shop --framework gin --db sqlite --data sqlc",
		},
	}
	for _, tt := range tests {
		if got := equivalentCommand(testSupported, tt.in, tt.f); got != tt.want {
			t.Errorf("%s:\ngot  %s\nwant %s", tt.name, got, tt.want)
		}
	}
}
