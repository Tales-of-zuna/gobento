package compose

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckRelPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dir     string
		wantErr string
	}{
		{dir: "internal/store"},
		{dir: "cmd/shop"},
		{dir: "", wantErr: "must not be empty"},
		{dir: ".", wantErr: "must not be empty"},
		{dir: `internal\store`, wantErr: "forward slashes"},
		{dir: "/internal/store", wantErr: "relative"},
		{dir: "internal//store", wantErr: "clean"},
		{dir: "internal/store/", wantErr: "clean"},
		{dir: "internal/../store", wantErr: "clean"},
		{dir: "..", wantErr: "inside the project"},
		{dir: "../store", wantErr: "inside the project"},
	}
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			t.Parallel()

			err := checkRelPath(tt.dir)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("checkRelPath(%q) error = %v, want nil", tt.dir, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("checkRelPath(%q) error = %v, want it to contain %q", tt.dir, err, tt.wantErr)
			}
		})
	}
}

func TestComposeSlotDirectoryErrors(t *testing.T) {
	t.Parallel()

	arch := baseFixture()["architecture/layered/component.yaml"]
	tests := []struct {
		name     string
		manifest string
		wantIs   error
		wantMsg  string
	}{
		{
			name:     "two slots share a directory",
			manifest: strings.Replace(arch, "dir: internal/store,", "dir: internal/database,", 1),
			wantIs:   ErrSlotOwner,
			wantMsg:  "slots database and store share the directory internal/database",
		},
		{
			name:     "directory leaves the project",
			manifest: strings.Replace(arch, "dir: internal/store,", "dir: ../store,", 1),
			wantIs:   ErrTemplate,
			wantMsg:  `slot store: dir "../store": must stay inside the project`,
		},
		{
			name:     "directory template fails",
			manifest: strings.Replace(arch, `"cmd/{{.Project.Name}}"`, `"cmd/{{.Project.Nam}}"`, 1),
			wantIs:   ErrTemplate,
			wantMsg:  "slot entry: dir:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cat := loadCatalog(t, withFiles(map[string]string{"architecture/layered/component.yaml": tt.manifest}))
			_, err := Compose(input(pick(t, cat, "layered", "web", "lite", "query")))
			if !errors.Is(err, tt.wantIs) {
				t.Fatalf("Compose() error = %v, want errors.Is(%v)", err, tt.wantIs)
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("Compose() error = %q\nwant it to contain %q", err, tt.wantMsg)
			}
		})
	}
}
