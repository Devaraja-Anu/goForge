// Package generator turns the embedded blueprint source tree into a real,
// standalone Go project on disk.
package generator

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// blueprintModulePath must match blueprint/go.mod's module line exactly.
const blueprintModulePath = "github.com/devaraja-anu/blueprint"

var goVersionPattern = regexp.MustCompile(`(?m)^go (\d+\.\d+(?:\.\d+)?)`)

// ciWorkflowFS embeds the generated project's fresh CI workflow template.
// Not sourced from blueprintsrc/ — see writeCIWorkflow's doc comment.
//
//go:embed templates/ci.yml.tmpl
var ciWorkflowFS embed.FS

type Options struct {
	ProjectName string
	ModulePath  string
	OutputDir   string
}

// excluded lists blueprintsrc/-relative paths never written to a generated
// project. See Decisions.md for the reasoning behind each entry.
var excluded = []string{
	".golangci.yml",
	"internal/api/helpers_test.go.tmpl",
	"internal/api/response_test.go.tmpl",
}

// Generate is all-or-nothing: on any failure, no partial output is left
// behind at OutputDir.
func Generate(src fs.FS, opts Options) error {
	if opts.ModulePath == "" {
		return fmt.Errorf("generator: ModulePath is required")
	}
	if _, err := os.Stat(opts.OutputDir); err == nil {
		return fmt.Errorf("generator: %s already exists", opts.OutputDir)
	}

	// tmpDir must be on the same volume as OutputDir, or the final
	// os.Rename silently degrades from an atomic move to copy+delete.
	parent := filepath.Dir(opts.OutputDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("generator: preparing output parent: %w", err)
	}

	tmpDir, err := os.MkdirTemp(parent, ".goforge-new-*")
	if err != nil {
		return fmt.Errorf("generator: creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir) // no-op once renamed away on success

	var goModContent []byte

	err = fs.WalkDir(src, "blueprintsrc", func(walkPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		content, err := processFile(src, walkPath, tmpDir, opts.ModulePath)
		if err != nil {
			return err
		}

		if strings.TrimPrefix(walkPath, "blueprintsrc/") == "go.mod.tmpl" {
			goModContent = content
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("generator: %w", err)
	}

	if err := writeCIWorkflow(tmpDir, goModContent); err != nil {
		return fmt.Errorf("generator: %w", err)
	}

	if err := os.Rename(tmpDir, opts.OutputDir); err != nil {
		return fmt.Errorf("generator: finalizing output: %w", err)
	}

	return nil
}

func processFile(src fs.FS, walkPath, tmpDir, newModulePath string) ([]byte, error) {
	relPath := strings.TrimPrefix(walkPath, "blueprintsrc/")
	if isExcluded(relPath) {
		return nil, nil
	}

	content, err := fs.ReadFile(src, walkPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", walkPath, err)
	}
	if needsRewrite(relPath) {
		content = rewriteModulePath(content, newModulePath)
	}

	outRelPath := outputPath(relPath)
	outFullPath := filepath.Join(tmpDir, outRelPath)
	if err := os.MkdirAll(filepath.Dir(outFullPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating dir for %s: %w", outRelPath, err)
	}
	if err := os.WriteFile(outFullPath, content, 0o644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", outRelPath, err)
	}
	return content, nil
}

func isExcluded(relPath string) bool {
	if strings.HasPrefix(relPath, ".github/") {
		return true
	}
	for _, e := range excluded {
		if relPath == e {
			return true
		}
	}
	return false
}

// outputPath reverses the .tmpl suffixes scripts/sync-blueprint.sh adds.
func outputPath(relPath string) string {
	switch {
	case relPath == "go.mod.tmpl":
		return "go.mod"
	case relPath == "go.sum.tmpl":
		return "go.sum"
	case strings.HasSuffix(relPath, ".go.tmpl"):
		return strings.TrimSuffix(relPath, ".tmpl")
	default:
		return relPath
	}
}

// needsRewrite excludes go.sum.tmpl deliberately: it's keyed by
// third-party dependency paths, not blueprint's own module path.
func needsRewrite(relPath string) bool {
	return strings.HasSuffix(relPath, ".go.tmpl") || relPath == "go.mod.tmpl"
}

func rewriteModulePath(content []byte, newModulePath string) []byte {
	return bytes.ReplaceAll(content, []byte(blueprintModulePath), []byte(newModulePath))
}

func extractGoVersion(goModContent []byte) (string, error) {
	m := goVersionPattern.FindSubmatch(goModContent)
	if m == nil {
		return "", fmt.Errorf("no 'go' directive found in go.mod")
	}
	return string(m[1]), nil
}

// writeCIWorkflow generates .github/workflows/ci.yml fresh, using the Go
// version pinned in the (already-processed) go.mod content from the embedded ci.yml file
func writeCIWorkflow(tmpDir string, goModContent []byte) error {
	goVersion, err := extractGoVersion(goModContent)
	if err != nil {
		return fmt.Errorf("determining Go version for CI workflow: %w", err)
	}

	tmplContent, err := fs.ReadFile(ciWorkflowFS, "templates/ci.yml.tmpl")
	if err != nil {
		return fmt.Errorf("reading CI workflow template: %w", err)
	}

	tmpl, err := template.New("ci").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("parsing CI workflow template: %w", err)
	}

	outPath := filepath.Join(tmpDir, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("creating .github/workflows: %w", err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating ci.yml: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, struct{ GoVersion string }{GoVersion: goVersion})
}
