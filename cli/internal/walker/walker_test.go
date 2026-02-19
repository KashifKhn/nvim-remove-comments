package walker

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestWalk_ReturnsOnlySupportedExtensions(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"main.go":       "package main",
		"script.py":     "x = 1",
		"README.md":     "# readme",
		"data.json":     "{}",
		"app.ts":        "const x = 1",
		"index.html":    "<html></html>",
		"style.css":     "body {}",
		"config.yaml":   "key: val",
		"config.yml":    "key: val",
		"settings.toml": "[settings]",
		"run.sh":        "echo hi",
		"Main.java":     "class Main {}",
		"main.c":        "int main(){}",
		"main.cpp":      "int main(){}",
		"main.rs":       "fn main(){}",
		"main.lua":      "local x = 1",
		"component.jsx": "const C = () => <div/>",
		"component.tsx": "const C = () => <div/>",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	entries, errs := Walk(dir, "", 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	unsupported := map[string]bool{".md": true, ".json": true}
	for _, e := range entries {
		if unsupported[e.Ext] {
			t.Errorf("unexpected file with extension %s: %s", e.Ext, e.Path)
		}
	}

	extsSeen := map[string]bool{}
	for _, e := range entries {
		extsSeen[e.Ext] = true
	}

	expected := []string{".go", ".py", ".ts", ".html", ".css", ".yaml", ".yml", ".toml", ".sh", ".java", ".c", ".cpp", ".rs", ".lua", ".jsx", ".tsx"}
	for _, ext := range expected {
		if !extsSeen[ext] {
			t.Errorf("expected extension %s not found in results", ext)
		}
	}
}

func TestWalk_RespectsGitignore(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("ignored.go\nbuild/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	buildDir := filepath.Join(dir, "build")
	if err := os.Mkdir(buildDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "output.go"), []byte("package build"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, errs := Walk(dir, "", 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, e := range entries {
		base := filepath.Base(e.Path)
		if base == "ignored.go" {
			t.Errorf("ignored.go should have been excluded by .gitignore")
		}
		if filepath.Dir(e.Path) == buildDir {
			t.Errorf("build/ should have been excluded by .gitignore")
		}
	}

	found := false
	for _, e := range entries {
		if filepath.Base(e.Path) == "main.go" {
			found = true
		}
	}
	if !found {
		t.Error("main.go should have been included")
	}
}

func TestWalk_LangFilter(t *testing.T) {
	dir := t.TempDir()

	filesToCreate := []string{"main.go", "script.py", "app.ts", "lib.rs"}
	for _, f := range filesToCreate {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	entries, errs := Walk(dir, "go", 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, e := range entries {
		if e.Lang.Name != "go" {
			t.Errorf("lang filter 'go' should not return file: %s (lang=%s)", e.Path, e.Lang.Name)
		}
	}
	if len(entries) == 0 {
		t.Error("expected at least one go file")
	}
}

func TestWalk_MaxFileSizeExcludes(t *testing.T) {
	dir := t.TempDir()

	small := filepath.Join(dir, "small.go")
	large := filepath.Join(dir, "large.go")

	if err := os.WriteFile(small, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	bigContent := make([]byte, 200)
	for i := range bigContent {
		bigContent[i] = 'x'
	}
	if err := os.WriteFile(large, bigContent, 0644); err != nil {
		t.Fatal(err)
	}

	entries, errs := Walk(dir, "", 100)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, e := range entries {
		if filepath.Base(e.Path) == "large.go" {
			t.Error("large.go should have been excluded by max-file-size")
		}
	}

	found := false
	for _, e := range entries {
		if filepath.Base(e.Path) == "small.go" {
			found = true
		}
	}
	if !found {
		t.Error("small.go should have been included")
	}
}

func TestWalk_SingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, errs := Walk(path, "", 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != path {
		t.Errorf("expected path %s, got %s", path, entries[0].Path)
	}
}

func TestWalk_SingleFileUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "README.md")
	if err := os.WriteFile(path, []byte("# hi"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, errs := Walk(path, "", 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for unsupported extension, got %d", len(entries))
	}
}

func TestWalk_PathDoesNotExist(t *testing.T) {
	_, errs := Walk("/nonexistent/path/that/does/not/exist", "", 0)
	if len(errs) == 0 {
		t.Error("expected an error for non-existent path")
	}
}

func TestWalk_NestedGitignore(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(subDir, ".gitignore"), []byte("secret.go\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "secret.go"), []byte("package sub"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "public.go"), []byte("package sub"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "root.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, errs := Walk(dir, "", 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, filepath.Base(e.Path))
	}
	sort.Strings(names)

	for _, n := range names {
		if n == "secret.go" {
			t.Error("secret.go should have been excluded by nested .gitignore")
		}
	}

	foundPublic := false
	for _, n := range names {
		if n == "public.go" {
			foundPublic = true
		}
	}
	if !foundPublic {
		t.Error("public.go should have been included")
	}
}

func TestWalk_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	entries, errs := Walk(dir, "", 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty dir, got %d", len(entries))
	}
}

func TestWalk_ExcludesNodeModules(t *testing.T) {
	dir := t.TempDir()
	nmDir := filepath.Join(dir, "node_modules")
	if err := os.Mkdir(nmDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nmDir, "dep.js"), []byte("// dep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("const x = 1"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, errs := Walk(dir, "", 0)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	for _, e := range entries {
		if filepath.Dir(e.Path) == nmDir {
			t.Errorf("node_modules should be excluded: %s", e.Path)
		}
	}
}
