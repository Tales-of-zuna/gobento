package generate

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestAssetSources covers the spec scenario "Vendored file matches
// upstream": for every asset in the shipped rules file, the file at its
// path inside the source archive has the recorded checksum. Together with
// compose's TestShippedAssets, which compares the checksum with the shipped
// file, it proves that the shipped file is the upstream one. It needs
// network access and is skipped with -short.
func TestAssetSources(t *testing.T) {
	if testing.Short() {
		t.Skip("downloads the upstream release archives")
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	for _, a := range shippedCatalog(t).Rules().Assets {
		t.Run(a.Name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, a.Source, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("download %s: %v", a.Source, err)
			}
			defer resp.Body.Close() //nolint:errcheck // read-only
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("download %s: %s", a.Source, resp.Status)
			}
			sum, err := fileSum(resp.Body, a.Path)
			if err != nil {
				t.Fatalf("%s: %v", a.Source, err)
			}
			if sum != a.SHA256 {
				t.Errorf("%s in %s has sha256 %s, but versions.yaml records %s", a.Path, a.Source, sum, a.SHA256)
			}
		})
	}
}

// fileSum returns the SHA-256 checksum of the file name in the gzipped tar
// archive r.
func fileSum(r io.Reader, name string) (string, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return "", errors.New(name + " is not in the archive")
		}
		if err != nil {
			return "", err
		}
		if h.Name != name {
			continue
		}
		sum := sha256.New()
		if _, err := io.Copy(sum, tr); err != nil { //nolint:gosec // G110: a small, pinned upstream file
			return "", err
		}
		return hex.EncodeToString(sum.Sum(nil)), nil
	}
}
