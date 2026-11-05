package generate

import (
	"bufio"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestBinaryChecksums covers the spec scenario "Checksum copied wrongly":
// for every binary in the shipped rules file that names a checksum file,
// every platform's SHA-256 equals the one the release publishes for that
// platform's file, and every download URL exists. It needs network access
// and is skipped with -short; it downloads the checksum file only, not the
// binaries, which the integration test downloads and checks for its own
// platform.
func TestBinaryChecksums(t *testing.T) {
	if testing.Short() {
		t.Skip("downloads the release checksum files")
	}
	client := &http.Client{Timeout: time.Minute}
	get := func(t *testing.T, method, url string) *http.Response {
		t.Helper()
		req, err := http.NewRequestWithContext(t.Context(), method, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", method, url, err)
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			t.Fatalf("%s %s: %s", method, url, resp.Status)
		}
		return resp
	}
	for _, b := range shippedCatalog(t).Rules().Binaries {
		t.Run(b.Name, func(t *testing.T) {
			if b.Checksums == "" {
				t.Skip("the release publishes no checksum file")
			}
			resp := get(t, http.MethodGet, b.Checksums)
			sums, err := parseChecksums(resp.Body)
			_ = resp.Body.Close()
			if err != nil {
				t.Fatalf("%s: %v", b.Checksums, err)
			}
			for _, p := range b.Platforms {
				name := p.FileName()
				switch got, ok := sums[name]; {
				case !ok:
					t.Errorf("%s: %s lists no %s", p.Key(), b.Checksums, name)
				case got != p.SHA256:
					t.Errorf("%s: %s lists sha256 %s for %s, but versions.yaml records %s", p.Key(), b.Checksums, got, name, p.SHA256)
				}
				head := get(t, http.MethodHead, p.URL)
				_ = head.Body.Close()
			}
		})
	}
}

// parseChecksums reads a checksum file of "<sha256>  <name>" lines, as
// sha256sum writes them, and returns the checksums keyed by file name
// without a leading "./" or "*".
func parseChecksums(r io.Reader) (map[string]string, error) {
	sums := make(map[string]string)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 2 {
			continue
		}
		name := strings.TrimPrefix(strings.TrimPrefix(fields[1], "*"), "./")
		sums[name] = fields[0]
	}
	return sums, sc.Err()
}
