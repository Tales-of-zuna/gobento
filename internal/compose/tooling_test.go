package compose

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
	"github.com/Tales-of-zuna/gobento/internal/compat"
)

// toolingFiles are the files the shipped base layer produces.
var toolingFiles = []string{
	".dockerignore",
	".github/dependabot.yml",
	".github/workflows/ci.yml",
	".golangci.yml",
	"Dockerfile",
	"Makefile",
}

// TestShippedToolingFiles covers the spec scenario "Files for every
// architecture": every supported selection gets the base layer's files.
func TestShippedToolingFiles(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	for _, sel := range Supported(cat) {
		plan := planFor(t, cat, sel)
		for _, p := range toolingFiles {
			f, ok := plan.File(p)
			if !ok {
				t.Errorf("%s: %s is not planned", sel, p)
				continue
			}
			if f.Component != catalog.BaseID {
				t.Errorf("%s: %s comes from %s, want the base layer", sel, p, f.Component)
			}
		}
	}
}

// TestShippedToolingFollowsDatabase covers the spec scenarios "Database
// tests run in CI", "No server for file databases", and "Writable data for
// file databases in the image".
func TestShippedToolingFollowsDatabase(t *testing.T) {
	t.Parallel()

	sqlite, postgres := shippedPlan(t, "sqlite"), shippedPlan(t, "postgres")
	file := func(p *Plan, path string) string {
		t.Helper()
		f, ok := p.File(path)
		if !ok {
			t.Fatalf("%s is not planned", path)
		}
		return string(f.Content)
	}

	pgCI := file(postgres, ".github/workflows/ci.yml")
	for _, want := range []string{
		"    services:\n      postgres:\n        image: postgres:",
		"    env:\n",
		"      TEST_DATABASE_URL: postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable\n",
	} {
		if !strings.Contains(pgCI, want) {
			t.Errorf("postgres workflow lacks %q:\n%s", want, pgCI)
		}
	}
	liteCI := file(sqlite, ".github/workflows/ci.yml")
	for _, unwanted := range []string{"services:", "TEST_DATABASE_URL"} {
		if strings.Contains(liteCI, unwanted) {
			t.Errorf("sqlite workflow contains %q:\n%s", unwanted, liteCI)
		}
	}

	liteDocker := file(sqlite, "Dockerfile")
	for _, want := range []string{"RUN mkdir -p /out/data", "COPY --from=build --chown=65532:65532 /out/data /data", "ENV DATABASE_URL=/data/shop.db", "VOLUME /data"} {
		if !strings.Contains(liteDocker, want) {
			t.Errorf("sqlite Dockerfile lacks %q", want)
		}
	}
	pgDocker := file(postgres, "Dockerfile")
	if strings.Contains(pgDocker, "/data") {
		t.Errorf("postgres Dockerfile has a data volume:\n%s", pgDocker)
	}
	for _, p := range []*Plan{sqlite, postgres} {
		d := file(p, "Dockerfile")
		for _, want := range []string{"CGO_ENABLED=0 go build", "USER 65532:65532", "FROM gcr.io/distroless/"} {
			if !strings.Contains(d, want) {
				t.Errorf("Dockerfile lacks %q", want)
			}
		}
	}
	if !strings.Contains(file(sqlite, ".dockerignore"), "*.db\n") || strings.Contains(file(postgres, ".dockerignore"), "*.db") {
		t.Error("only the sqlite .dockerignore should ignore *.db files")
	}
	if !strings.Contains(file(postgres, "README.md"), "host.docker.internal") {
		t.Error("postgres README does not explain how the container reaches the server")
	}
}

// TestShippedToolingPinsActions covers the spec scenario "Actions pinned by
// SHA" and "Ecosystems".
func TestShippedToolingPinsActions(t *testing.T) {
	t.Parallel()

	plan := shippedPlan(t, "postgres")
	ci, _ := plan.File(".github/workflows/ci.yml")
	uses := regexp.MustCompile(`(?m)^\s*(?:- )?uses: (.*)$`).FindAllStringSubmatch(string(ci.Content), -1)
	if len(uses) == 0 {
		t.Fatal("workflow uses no actions")
	}
	pinned := regexp.MustCompile(`^[\w.-]+/[\w./-]+@[0-9a-f]{40} # v\d+\.\d+\.\d+$`)
	for _, u := range uses {
		if !pinned.MatchString(u[1]) {
			t.Errorf("uses: %s is not pinned by commit SHA with a version comment", u[1])
		}
	}

	dependabot, _ := plan.File(".github/dependabot.yml")
	for _, eco := range []string{"gomod", "github-actions", "docker"} {
		if !strings.Contains(string(dependabot.Content), "package-ecosystem: "+eco+"\n") {
			t.Errorf("dependabot.yml lacks the %s ecosystem", eco)
		}
	}
}

// TestShippedToolingGoImage covers the spec scenario "Go version follows the
// selection".
func TestShippedToolingGoImage(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	sel := pick(t, cat, "mvc", "nethttp", "sqlite", "sqlc")
	var ids []string
	for _, c := range sel.Components() {
		ids = append(ids, c.ID())
	}
	res, err := cat.Rules().Resolve(compat.Request{Components: ids, Traits: sel.Traits(), Go: "1.26"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Compose(Input{
		Project:   Project{Name: "shop", Module: "example.com/shop", GoVersion: "1.26.0"},
		Selection: sel,
		Versions:  res.Versions(),
		Base:      cat.Base(),
		Tooling:   res.Tooling,
	})
	if err != nil {
		t.Fatal(err)
	}
	d, _ := plan.File("Dockerfile")
	if !strings.Contains(string(d.Content), "\nFROM golang:1.26 AS build\n") {
		t.Errorf("Dockerfile does not build with golang:1.26:\n%s", d.Content)
	}
}

// TestShippedTemplatesHaveNoImageVersions checks the rule "Templates SHALL
// NOT contain version numbers" for the tooling: no template names an image
// tag, an action version, or a commit SHA itself.
func TestShippedTemplatesHaveNoImageVersions(t *testing.T) {
	t.Parallel()

	cat := shippedCatalog(t)
	forbidden := regexp.MustCompile(`(?:postgres|golang|distroless/[\w-]+):\d|@v\d|[0-9a-f]{40}|sha256:`)
	comps := []*catalog.Component{cat.Base()}
	for _, k := range catalog.Kinds() {
		comps = append(comps, cat.Components(k)...)
	}
	for _, c := range comps {
		texts := map[string]string{"notes": c.Notes}
		for _, f := range c.Files {
			texts[f.Source] = f.Template
		}
		for point, text := range c.Contributions {
			texts["contrib/"+point] = text
		}
		for name, text := range texts {
			// sqlcgen holds sqlc's own version comment, which is part of
			// the generated code it ships.
			if strings.Contains(name, "sqlcgen/") {
				continue
			}
			if m := forbidden.FindString(text); m != "" {
				t.Errorf("%s: %s contains the version %q; take it from versions.yaml", c.ID(), name, m)
			}
		}
	}
}
