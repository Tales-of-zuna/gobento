// Command tailwindcss runs the standalone Tailwind CSS CLI 4.3.4, the
// version this project's stylesheet is built with, so that no Node.js is
// needed. It passes its arguments to the CLI:
//
//	go run ./internal/tools/tailwindcss --input tailwind.css --output internal/views/static/styles.css --minify
//
// On first use it downloads the CLI for the current platform from the
// Tailwind CSS release, checks its SHA-256, and caches it in the user cache
// directory. Set TAILWINDCSS to the path of a tailwindcss binary to run that
// one instead, for example on a platform the release has no build for.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"
)

// version is the Tailwind CSS release. After changing it and releases,
// rebuild the stylesheet and commit both.
const version = "4.3.4"

// release is the file of the release for one platform.
type release struct {
	url string
	sum string // SHA-256 in lowercase hex
}

// releases maps GOOS/GOARCH to the file of the release for that platform.
var releases = map[string]release{
	"darwin/amd64":  {url: "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.4/tailwindcss-macos-x64", sum: "7ec771593a1fb8bbd204f70429361b0158f7989e4136bc68d04513bf543a15ba"},
	"darwin/arm64":  {url: "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.4/tailwindcss-macos-arm64", sum: "a516a11b439bfbb5ecd4a2d13b75e6cdcc284dc710aed8731e31d2e5c03895a4"},
	"linux/amd64":   {url: "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.4/tailwindcss-linux-x64", sum: "149a8ece5b855587a9c10de6da9dcbcfad0e47c4ed5dca219f721b3eae39d817"},
	"linux/arm64":   {url: "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.4/tailwindcss-linux-arm64", sum: "f6d7baae1d350ed3a2defe98d0b5cc58e37400c0173b2e38d42f690ccc4bbfe1"},
	"windows/amd64": {url: "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.4/tailwindcss-windows-x64.exe", sum: "ad84612696f3be102fa005d4827f0df84c7b4570d31c06865ce5fde0311478af"},
	"windows/arm64": {url: "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.4/tailwindcss-windows-x64.exe", sum: "ad84612696f3be102fa005d4827f0df84c7b4570d31c06865ce5fde0311478af"},
}

func main() {
	err := run(os.Args[1:])
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		os.Exit(exit.ExitCode())
	case err != nil:
		fmt.Fprintln(os.Stderr, "tailwindcss:", err)
		os.Exit(1)
	}
}

// run runs the CLI with args, connected to this process's standard streams.
func run(args []string) error {
	bin, err := binary(context.Background(), runtime.GOOS+"/"+runtime.GOARCH)
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...) //nolint:gosec // G204: the verified CLI, or the one TAILWINDCSS names
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// binary returns the path of the CLI for platform: $TAILWINDCSS, or the
// cached download, which it downloads first if it is missing or does not
// have the expected checksum.
func binary(ctx context.Context, platform string) (string, error) {
	if p := os.Getenv("TAILWINDCSS"); p != "" {
		return p, nil
	}
	rel, ok := releases[platform]
	if !ok {
		return "", fmt.Errorf("the Tailwind CSS %s release has no build for %s (it has builds for %s); set TAILWINDCSS to the path of a tailwindcss %s binary",
			version, platform, strings.Join(slices.Sorted(maps.Keys(releases)), ", "), version)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("find the cache directory: %w; set TAILWINDCSS to the path of a tailwindcss %s binary", err, version)
	}
	dir := filepath.Join(cache, "tailwindcss", version)
	bin := filepath.Join(dir, path.Base(rel.url))
	if sum, err := fileSHA256(bin); err == nil && sum == rel.sum {
		return bin, nil
	}
	if err := download(ctx, rel, bin); err != nil {
		return "", err
	}
	return bin, nil
}

// download saves the file of rel as bin, an executable, if its checksum
// matches; otherwise it keeps nothing.
func download(ctx context.Context, rel release, bin string) (err error) {
	if err := os.MkdirAll(filepath.Dir(bin), 0o750); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Downloading Tailwind CSS %s from %s\n", version, rel.url)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rel.url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download Tailwind CSS: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // read-only
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", rel.url, resp.Status)
	}

	tmp, err := os.CreateTemp(filepath.Dir(bin), ".download-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()
	sum := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, sum), resp.Body); err != nil {
		return fmt.Errorf("download %s: %w", rel.url, err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if got := hex.EncodeToString(sum.Sum(nil)); got != rel.sum {
		return fmt.Errorf("%s has SHA-256 %s, want %s; the download was deleted", rel.url, got, rel.sum)
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil { //nolint:gosec // G302: the CLI must be executable
		return err
	}
	// A stale file must go first: Windows cannot rename onto an existing file.
	if err := os.Remove(bin); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(tmp.Name(), bin)
}

// fileSHA256 returns the SHA-256 checksum of the file at name in lowercase
// hex.
func fileSHA256(name string) (string, error) {
	f, err := os.Open(name) //nolint:gosec // G304: the cached CLI's path
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint:errcheck // read-only
	sum := sha256.New()
	if _, err := io.Copy(sum, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}
