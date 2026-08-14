package architecture_test

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestAgentInstructionsMatch(t *testing.T) {
	root := moduleRoot(t)
	claude, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(agents, claude) {
		t.Fatal("AGENTS.md and CLAUDE.md must be byte-identical")
	}
	workflow, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(workflow), "cmp --silent AGENTS.md CLAUDE.md") {
		t.Fatal("CI must block when AGENTS.md and CLAUDE.md differ")
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestEntrypointsStayThin(t *testing.T) {
	root := moduleRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "cmd", "service", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(string(data), "\n") + 1; lines > 30 {
		t.Fatalf("cmd/service/main.go has %d lines; move composition into internal/di", lines)
	}
	forbidden := []string{"internal/config", "pkg/postgres", "pkg/service", "log.Fatal"}
	for _, value := range forbidden {
		if strings.Contains(string(data), value) {
			t.Errorf("cmd/service/main.go contains %q", value)
		}
	}
}

func TestImplementationDoesNotReturnToModuleRoot(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(moduleRoot(t), "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("root-level Go files are forbidden: %v", files)
	}
}

func TestEveryHandlerOperationHasItsOwnTest(t *testing.T) {
	handlerRoot := filepath.Join(moduleRoot(t), "internal", "handler")
	entries, err := os.ReadDir(handlerRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "validator" {
			continue
		}
		dir := filepath.Join(handlerRoot, entry.Name())
		if _, err := os.Stat(filepath.Join(dir, "handler.go")); err != nil {
			t.Errorf("%s: missing handler.go", entry.Name())
		}
		if _, err := os.Stat(filepath.Join(dir, "handler_test.go")); err != nil {
			t.Errorf("%s: missing handler_test.go", entry.Name())
		}
	}
}

func TestDependencyDirection(t *testing.T) {
	root := moduleRoot(t)
	checkImports(t, filepath.Join(root, "internal", "model"), []string{
		"net/http", "database/sql", "/internal/handler", "/internal/di", "/internal/pkg/",
	})
	checkImports(t, filepath.Join(root, "internal", "pkg", "usecase"), []string{
		"net/http", "database/sql", "github.com/jackc/pgx", "/internal/handler", "/internal/di",
		"/internal/pkg/clients", "/internal/pkg/storage",
	})
}

func TestLintVersionMatchesCI(t *testing.T) {
	root := moduleRoot(t)
	makefile, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	version := regexp.MustCompile(`(?m)^GOLANGCI_LINT_VERSION := (v\d+\.\d+\.\d+)$`).FindSubmatch(makefile)
	if len(version) != 2 {
		t.Fatal("Makefile must declare an exact GOLANGCI_LINT_VERSION")
	}
	if !strings.Contains(string(makefile), "golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh") {
		t.Error("Makefile must download the installer from the pinned golangci-lint tag")
	}
	if strings.Contains(string(makefile), "golangci-lint/master/") {
		t.Error("Makefile must not download golangci-lint from master")
	}
	if !strings.Contains(string(workflow), "version: "+string(version[1])) {
		t.Errorf("CI must use the Makefile golangci-lint version %s", version[1])
	}
}

func checkImports(t *testing.T, root string, forbidden []string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			value := strings.Trim(spec.Path.Value, `"`)
			for _, denied := range forbidden {
				if value == denied || strings.Contains(value, denied) {
					t.Errorf("%s imports forbidden dependency %s", path, value)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
