package generate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// The fake node and npm commands, configured by these environment
// variables:
//
//	FAKE_NODE_VERSION    output of node --version (default v24.22.0)
//	FAKE_NPM_LOG         file that receives each npm invocation
//	FAKE_NPM=fail        npm create fails as if offline
//	FAKE_NPM=hang        npm create starts a child process, creates
//	                     FAKE_NPM_STARTED, and blocks; the child creates
//	                     FAKE_NPM_LATE after two seconds unless killed
//	FAKE_NPM_NO_REMOVE=1 npm create does not create src/App.css
//
// npm create writes a small project with package.json, src/App.tsx,
// src/App.css, and src/assets/logo.svg; npm install writes
// package-lock.json listing its arguments.
func fakeNode(args []string) int {
	if len(args) == 1 && args[0] == "--version" {
		fmt.Println(envOr("FAKE_NODE_VERSION", "v24.22.0"))
		return 0
	}
	fmt.Fprintf(os.Stderr, "fake node: unexpected arguments %q\n", args)
	return 2
}

func fakeNPM(args []string) int {
	if os.Getenv("FAKE_NPM_ROLE") == "child" {
		time.Sleep(2 * time.Second)
		_ = os.WriteFile(os.Getenv("FAKE_NPM_LATE"), []byte("late"), 0o600) //nolint:gosec // G703: test-controlled path
		return 0
	}
	if log := os.Getenv("FAKE_NPM_LOG"); log != "" {
		f, err := os.OpenFile(log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // G304: test-controlled path
		if err == nil {
			wd, _ := os.Getwd()
			_, _ = fmt.Fprintf(f, "npm %s | dir=%s yes=%s\n", strings.Join(args, " "), filepath.Base(wd), os.Getenv("npm_config_yes"))
			_ = f.Close()
		}
	}
	switch {
	case len(args) > 0 && args[0] == "create":
		return fakeCreate(args[1:])
	case len(args) > 0 && args[0] == "install":
		data := "{\"fake\": \"" + strings.Join(args[1:], " ") + "\"}\n"
		if err := os.WriteFile("package-lock.json", []byte(data), 0o600); err != nil { //nolint:gosec // G703: fixed file name in the fake's working directory
			fmt.Fprintln(os.Stderr, "fake npm:", err)
			return 1
		}
		return 0
	}
	fmt.Fprintf(os.Stderr, "fake npm: unexpected arguments %q\n", args)
	return 2
}

// fakeCreate handles npm create --yes <pkg>@<ver> <dir> -- <flags>.
func fakeCreate(args []string) int {
	switch os.Getenv("FAKE_NPM") {
	case "fail":
		fmt.Fprintln(os.Stderr, "npm error code ENOTFOUND\nnpm error network request to https://registry.npmjs.org/create-vite failed")
		return 1
	case "hang":
		self, _ := os.Executable()
		child := exec.Command(self) //nolint:gosec // G204: the test binary itself
		child.Env = append(os.Environ(), "FAKE_NPM_ROLE=child")
		if err := child.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		_ = os.WriteFile(os.Getenv("FAKE_NPM_STARTED"), nil, 0o600) //nolint:gosec // G703: test-controlled path
		time.Sleep(time.Minute)
		return 1
	}
	if len(args) < 3 || args[0] != "--yes" {
		fmt.Fprintf(os.Stderr, "fake npm create: unexpected arguments %q\n", args)
		return 2
	}
	dir := args[2]
	files := map[string]string{
		"package.json":        `{"name": "` + dir + `"}`,
		"index.html":          "<!doctype html>",
		"src/App.tsx":         "scaffolded",
		"src/assets/logo.svg": "<svg/>",
	}
	if os.Getenv("FAKE_NPM_NO_REMOVE") == "" {
		files["src/App.css"] = "scaffolded"
	}
	for name, data := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil { //nolint:gosec // G703: a fixed name under the directory the test asked for
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if err := os.WriteFile(p, []byte(data), 0o600); err != nil { //nolint:gosec // G703: a fixed name under the directory the test asked for
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return 0
}
