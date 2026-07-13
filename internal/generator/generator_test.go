package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestIsExcluded(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{".golangci.yml", true},
		{"internal/api/helpers_test.go.tmpl", true},
		{"internal/api/response_test.go.tmpl", true},
		{".github/workflows/blueprint-ci.yml", true},
		{".github/dependabot.yml", true},
		{"internal/config/config_test.go.tmpl", false}, // kept, per Decisions.md
		{"cmd/api/main.go.tmpl", false},
		{"Makefile", false},
	}
	for _, c := range cases {
		if got := isExcluded(c.path); got != c.want {
			t.Errorf("isExcluded(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestOutputPath(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"go.mod.tmpl", "go.mod"},
		{"go.sum.tmpl", "go.sum"},
		{"internal/api/routes.go.tmpl", "internal/api/routes.go"},
		{"cmd/api/main.go.tmpl", "cmd/api/main.go"},
		{"Makefile", "Makefile"},
		{"migrations/000001_init.up.sql", "migrations/000001_init.up.sql"},
	}
	for _, c := range cases {
		if got := outputPath(c.path); got != c.want {
			t.Errorf("outputPath(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestNeedsRewrite(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"go.mod.tmpl", true},
		{"go.sum.tmpl", false}, // third-party hashes, not blueprint's own path
		{"internal/api/routes.go.tmpl", true},
		{"Makefile", false},
	}
	for _, c := range cases {
		if got := needsRewrite(c.path); got != c.want {
			t.Errorf("needsRewrite(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestRewriteModulePath(t *testing.T) {
	in := []byte(`import "github.com/devaraja-anu/blueprint/internal/api"`)
	want := `import "github.com/example/myapi/internal/api"`

	got := rewriteModulePath(in, "github.com/example/myapi")
	if string(got) != want {
		t.Errorf("rewriteModulePath() = %q, want %q", got, want)
	}
}

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"blueprintsrc/go.mod.tmpl": &fstest.MapFile{
			Data: []byte("module github.com/devaraja-anu/blueprint\n\ngo 1.26\n"),
		},
		"blueprintsrc/go.sum.tmpl": &fstest.MapFile{
			Data: []byte("github.com/go-chi/chi/v5 v5.0.0 h1:abc=\n"),
		},
		"blueprintsrc/cmd/api/main.go.tmpl": &fstest.MapFile{
			Data: []byte(`package main

import "github.com/devaraja-anu/blueprint/internal/api"

func main() { api.Run() }
`),
		},
		"blueprintsrc/Makefile": &fstest.MapFile{
			Data: []byte("build:\n\tgo build ./...\n"),
		},
		"blueprintsrc/.golangci.yml": &fstest.MapFile{
			Data: []byte("version: \"2\"\n"),
		},
		"blueprintsrc/.github/workflows/blueprint-ci.yml": &fstest.MapFile{
			Data: []byte("name: blueprint-ci\n"),
		},
		"blueprintsrc/internal/api/helpers_test.go.tmpl": &fstest.MapFile{
			Data: []byte("package api\n"),
		},
		"blueprintsrc/internal/config/config_test.go.tmpl": &fstest.MapFile{
			Data: []byte("package config\n"),
		},
	}
}

func TestGenerate_WritesExpectedFiles(t *testing.T) {
	out := filepath.Join(t.TempDir(), "myapi")

	err := Generate(testFS(), Options{
		ModulePath: "github.com/example/myapi",
		OutputDir:  out,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Renamed + rewritten
	got, err := os.ReadFile(filepath.Join(out, "cmd/api/main.go"))
	if err != nil {
		t.Fatalf("reading generated main.go: %v", err)
	}
	if want := "github.com/example/myapi/internal/api"; !strings.Contains(string(got), want) {
		t.Errorf("main.go missing rewritten import %q, got:\n%s", want, got)
	}
	if strings.Contains(string(got), "github.com/devaraja-anu/blueprint") {
		t.Errorf("main.go still contains old module path:\n%s", got)
	}

	// go.mod renamed + rewritten
	if _, err := os.Stat(filepath.Join(out, "go.mod.tmpl")); !os.IsNotExist(err) {
		t.Errorf("go.mod.tmpl should not exist in output, err = %v", err)
	}
	gotMod, err := os.ReadFile(filepath.Join(out, "go.mod"))
	if err != nil {
		t.Fatalf("reading generated go.mod: %v", err)
	}
	if !strings.Contains(string(gotMod), "module github.com/example/myapi") {
		t.Errorf("go.mod not rewritten, got:\n%s", gotMod)
	}

	// go.sum renamed, NOT rewritten (still has original content)
	gotSum, err := os.ReadFile(filepath.Join(out, "go.sum"))
	if err != nil {
		t.Fatalf("reading generated go.sum: %v", err)
	}
	if !strings.Contains(string(gotSum), "github.com/go-chi/chi/v5") {
		t.Errorf("go.sum content changed unexpectedly, got:\n%s", gotSum)
	}

	// CI workflow generated fresh, not copied from blueprintsrc/
	gotCI, err := os.ReadFile(filepath.Join(out, ".github/workflows/ci.yml"))
	if err != nil {
		t.Fatalf("reading generated ci.yml: %v", err)
	}
	if !strings.Contains(string(gotCI), "go-version: '1.26'") {
		t.Errorf("ci.yml missing expected go-version, got:\n%s", gotCI)
	}
	if strings.Contains(string(gotCI), "golangci-lint") {
		t.Errorf("ci.yml should not contain a lint step, got:\n%s", gotCI)
	}

	// Untouched filename, passthrough content
	if _, err := os.Stat(filepath.Join(out, "Makefile")); err != nil {
		t.Errorf("Makefile missing from output: %v", err)
	}

	// Excluded files must NOT appear
	mustNotExist := []string{
		".golangci.yml",
		".github/workflows/blueprint-ci.yml",
		"internal/api/helpers_test.go",
	}
	for _, p := range mustNotExist {
		if _, err := os.Stat(filepath.Join(out, p)); !os.IsNotExist(err) {
			t.Errorf("excluded file %q should not exist in output", p)
		}
	}

	// Kept test file
	if _, err := os.Stat(filepath.Join(out, "internal/config/config_test.go")); err != nil {
		t.Errorf("config_test.go should exist in output: %v", err)
	}
}

func TestGenerate_RequiresModulePath(t *testing.T) {
	err := Generate(testFS(), Options{OutputDir: filepath.Join(t.TempDir(), "myapi")})
	if err == nil {
		t.Fatal("expected error when ModulePath is empty, got nil")
	}
}

func TestGenerate_RefusesExistingOutputDir(t *testing.T) {
	out := t.TempDir() // TempDir() already creates and returns an existing dir

	err := Generate(testFS(), Options{
		ModulePath: "github.com/example/myapi",
		OutputDir:  out,
	})
	if err == nil {
		t.Fatal("expected error when OutputDir already exists, got nil")
	}
}

func TestGenerate_NoTempDirLeftOnSuccess(t *testing.T) {
	parent := t.TempDir()
	out := filepath.Join(parent, "myapi")

	if err := Generate(testFS(), Options{ModulePath: "github.com/example/myapi", OutputDir: out}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatalf("reading parent dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "myapi" {
		t.Errorf("expected only %q in parent dir, got %v", "myapi", entries)
	}
}

func TestExtractGoVersion(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "standard go.mod",
			content: "module github.com/example/myapi\n\ngo 1.26.2\n",
			want:    "1.26.2",
		},
		{
			name:    "two-part version",
			content: "module github.com/example/myapi\n\ngo 1.26\n",
			want:    "1.26",
		},
		{
			name:    "missing go directive",
			content: "module github.com/example/myapi\n",
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := extractGoVersion([]byte(c.content))
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("extractGoVersion() = %q, want %q", got, c.want)
			}
		})
	}
}
