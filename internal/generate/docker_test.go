package generate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Tales-of-zuna/gobento/internal/catalog"
)

// TestIntegrationDocker builds the container image of a SQLite and a
// PostgreSQL project with the generated Dockerfile, and runs the SQLite one.
// It covers the spec scenarios "Image builds and serves", "Image runs as
// non-root", "SQLite data in the container", and "PostgreSQL image builds".
// It needs network access and a Docker daemon that runs Linux containers,
// and is skipped otherwise and with -short.
func TestIntegrationDocker(t *testing.T) {
	skipIntegration(t)
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker command not found")
	}
	out, err := docker("info", "--format", "{{.OSType}}").Output()
	if err != nil || strings.TrimSpace(string(out)) != "linux" {
		t.Skipf("no Docker daemon running Linux containers (docker info: %q, %v)", strings.TrimSpace(string(out)), err)
	}

	cat := shippedCatalog(t)
	for _, tt := range []struct {
		sel [4]string
		run bool
	}{
		{sel: [4]string{"mvc", "nethttp", "sqlite", "sqlc"}, run: true},
		{sel: [4]string{"ddd", "gin", "postgres", "gorm"}},
	} {
		t.Run(strings.Join(tt.sel[:], "-"), func(t *testing.T) {
			dir := generateNamed(t, cat, tt.sel)
			image := fmt.Sprintf("gobento-test-%s-%d", strings.Join(tt.sel[:], "-"), time.Now().UnixNano())
			runIn(t, dir, "docker", "build", "--tag", image, ".")
			t.Cleanup(func() { _ = docker("image", "rm", "--force", image).Run() })

			if user := strings.TrimSpace(runIn(t, dir, "docker", "image", "inspect", "--format", "{{.Config.User}}", image)); user != "65532:65532" {
				t.Errorf("image user = %q, want 65532:65532", user)
			}
			if tt.run {
				checkContainer(t, dir, image)
			}
		})
	}
}

// docker returns a docker command with the given arguments.
func docker(args ...string) *exec.Cmd {
	return exec.Command("docker", args...) //nolint:gosec // G204: the test runs docker with image and container names it created
}

// generateNamed generates the named selection with default versions into a
// temporary directory and returns it.
func generateNamed(t *testing.T, cat *catalog.Catalog, sel [4]string) string {
	t.Helper()
	p, err := Prepare(cat, Spec{Name: "shop", Module: "example.com/shop", Components: map[catalog.Kind]string{
		catalog.KindArchitecture: sel[0], catalog.KindFramework: sel[1], catalog.KindDatabase: sel[2], catalog.KindData: sel[3],
	}})
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	dir := filepath.Join(t.TempDir(), "shop")
	if err := Generate(context.Background(), p.Request(dir)); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	return dir
}

// checkContainer starts image, creates and lists an item over HTTP, and
// checks that the SQLite database file is in the /data volume.
func checkContainer(t *testing.T, dir, image string) {
	t.Helper()
	id := strings.TrimSpace(runIn(t, dir, "docker", "run", "--detach", "--rm", "--publish", "127.0.0.1::8080", image))
	t.Cleanup(func() {
		if t.Failed() {
			logs, _ := docker("logs", id).CombinedOutput()
			t.Logf("container logs:\n%s", logs)
		}
		_ = docker("stop", id).Run()
	})
	addr := strings.TrimSpace(runIn(t, dir, "docker", "port", id, "8080/tcp"))
	addr = strings.Split(addr, "\n")[0]
	base := "http://" + addr

	client := &http.Client{Timeout: 5 * time.Second}
	request := func(method, path, body string) (int, string) {
		t.Helper()
		req, err := http.NewRequestWithContext(context.Background(), method, base+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return 0, err.Error()
		}
		defer func() { _ = resp.Body.Close() }()
		data, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(data)
	}

	deadline := time.Now().Add(30 * time.Second)
	for {
		status, body := request(http.MethodGet, "/health", "")
		if status == http.StatusOK {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET /health = %d %s, want 200", status, body)
		}
		time.Sleep(200 * time.Millisecond)
	}
	if status, body := request(http.MethodPost, "/items", `{"name":"boxed"}`); status != http.StatusCreated {
		t.Fatalf("POST /items = %d %s, want 201", status, body)
	}
	if status, body := request(http.MethodGet, "/items", ""); status != http.StatusOK || !strings.Contains(body, `"boxed"`) {
		t.Fatalf("GET /items = %d %s, want 200 with the created item", status, body)
	}

	// The database file is in the volume, written by the non-root user.
	copied := filepath.Join(t.TempDir(), "shop.db")
	runIn(t, dir, "docker", "cp", id+":/data/shop.db", copied)
	if info, err := os.Stat(copied); err != nil || info.Size() == 0 {
		t.Errorf("/data/shop.db in the container: %v, size %d", err, sizeOf(info))
	}
}

func sizeOf(info os.FileInfo) int64 {
	if info == nil {
		return 0
	}
	return info.Size()
}
