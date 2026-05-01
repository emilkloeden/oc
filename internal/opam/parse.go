package opam

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emilkloeden/oc/internal/project"
)

// FindOpamFile returns the path to the single *.opam file in dir.
// Returns an error if none or multiple are found.
func FindOpamFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read directory: %w", err)
	}
	var found []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".opam") {
			found = append(found, filepath.Join(dir, e.Name()))
		}
	}
	if len(found) == 0 {
		return "", fmt.Errorf("no .opam file found in %s", dir)
	}
	if len(found) > 1 {
		return "", fmt.Errorf("multiple .opam files found in %s", dir)
	}
	return found[0], nil
}

// AddDepToOpam adds a dependency to the depends: block in an opam file.
// If pkg is already present, this is a no-op.
func AddDepToOpam(path, pkg, constraint string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read opam file: %w", err)
	}
	content := string(data)

	if opamHasDep(content, pkg) {
		return nil
	}

	entry := formatOpamDepEntry(pkg, constraint)

	// Find the closing ']' of the depends: block and insert before it
	_, end, err := findOpamDepsBounds(content)
	if err != nil {
		return err
	}

	newContent := content[:end] + "  " + entry + "\n" + content[end:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

// RemoveDepFromOpam removes a dependency from the depends: block in an opam file.
// Returns an error if pkg is not found.
func RemoveDepFromOpam(path, pkg string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read opam file: %w", err)
	}
	content := string(data)

	if !opamHasDep(content, pkg) {
		return fmt.Errorf("%q is not in the depends: block", pkg)
	}

	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isOpamDepLine(trimmed, pkg) {
			continue
		}
		result = append(result, line)
	}
	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// findOpamDepsBounds finds the start and end positions of the depends: block.
// 'start' is after "depends: [", 'end' is the position of the closing ']'.
func findOpamDepsBounds(content string) (start, end int, err error) {
	idx := strings.Index(content, "depends: [")
	if idx < 0 {
		idx = strings.Index(content, "depends:[")
		if idx < 0 {
			return 0, 0, fmt.Errorf("no depends: field found in opam file")
		}
		start = idx + len("depends:[")
	} else {
		start = idx + len("depends: [")
	}

	end = strings.Index(content[start:], "]")
	if end < 0 {
		return 0, 0, fmt.Errorf("unterminated depends: block in opam file")
	}
	return start, start + end, nil
}

// opamHasDep reports whether pkg appears in the depends: block of an opam file.
func opamHasDep(content, pkg string) bool {
	start, end, err := findOpamDepsBounds(content)
	if err != nil {
		return false
	}
	block := content[start:end]
	for _, line := range strings.Split(block, "\n") {
		if isOpamDepLine(strings.TrimSpace(line), pkg) {
			return true
		}
	}
	return false
}

// isOpamDepLine reports whether a trimmed line is the dep entry for pkg.
// Opam dep lines look like: "pkg" or "pkg" {constraint}
func isOpamDepLine(trimmed, pkg string) bool {
	quoted := `"` + pkg + `"`
	return trimmed == quoted || strings.HasPrefix(trimmed, quoted+" ") || strings.HasPrefix(trimmed, quoted+"{")
}

// formatOpamDepEntry formats a dep entry for insertion into an opam depends: block.
func formatOpamDepEntry(pkg, constraint string) string {
	if constraint == "*" || constraint == "" {
		return fmt.Sprintf("%q", pkg)
	}
	op, ver := project.ParseConstraintParts(constraint)
	if op == "" {
		return fmt.Sprintf("%q", pkg)
	}
	return fmt.Sprintf("%q {%s %q}", pkg, op, ver)
}

// ReadOCamlVersion reads the OCaml version constraint from the .opam file in dir.
// It returns the version string from the first `"ocaml" {>= "x.y.z"}` or `{= "x.y.z"}` line.
func ReadOCamlVersion(dir string) (string, error) {
	path, err := FindOpamFile(dir)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read opam file: %w", err)
	}
	start, end, err := findOpamDepsBounds(string(data))
	if err != nil {
		return "", fmt.Errorf("find depends block: %w", err)
	}
	block := string(data)[start:end]
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, `"ocaml"`) {
			continue
		}
		// Extract version from {>= "x.y.z"} or {= "x.y.z"}
		open := strings.Index(trimmed, `"ocaml"`) + len(`"ocaml"`)
		rest := strings.TrimSpace(trimmed[open:])
		if !strings.HasPrefix(rest, "{") {
			continue
		}
		inner := strings.TrimPrefix(rest, "{")
		inner = strings.TrimSuffix(strings.TrimSpace(inner), "}")
		inner = strings.TrimSpace(inner)
		for _, op := range []string{">=", "<=", ">", "<", "="} {
			if strings.HasPrefix(inner, op) {
				ver := strings.TrimSpace(inner[len(op):])
				ver = strings.Trim(ver, `"`)
				if ver != "" {
					return ver, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no ocaml version constraint found in %s", path)
}
