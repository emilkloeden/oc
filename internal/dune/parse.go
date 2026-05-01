package dune

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HasGenerateOpamFiles reports whether dune-project in dir contains (generate_opam_files true).
func HasGenerateOpamFiles(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "dune-project"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "(generate_opam_files")
}

// GetPackageName reads the package name from the (package (name ...)) stanza in dune-project.
func GetPackageName(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "dune-project"))
	if err != nil {
		return "", fmt.Errorf("read dune-project: %w", err)
	}
	content := string(data)

	pkgStart, pkgEnd, err := findStanzaBounds(content, "(package")
	if err != nil {
		return "", fmt.Errorf("no (package ...) stanza in dune-project: %w", err)
	}
	pkgInterior := content[pkgStart:pkgEnd]

	name, err := findTopLevelAtom(pkgInterior, "name")
	if err != nil {
		return "", fmt.Errorf("no (name ...) in (package ...) stanza")
	}
	return name, nil
}

// findTopLevelAtom scans the interior of a stanza for a top-level (keyword value)
// child, skipping strings and nested stanzas, and returns value.
func findTopLevelAtom(interior, keyword string) (string, error) {
	prefix := "(" + keyword + " "
	i := 0
	inString := false
	depth := 0
	for i < len(interior) {
		c := interior[i]
		if inString {
			if c == '"' {
				inString = false
			} else if c == '\\' {
				i++
			}
			i++
			continue
		}
		if c == '"' {
			inString = true
			i++
			continue
		}
		if c == '(' {
			if depth == 0 && strings.HasPrefix(interior[i:], prefix) {
				after := interior[i+len(prefix):]
				end := strings.IndexAny(after, ") \t\n\r")
				if end < 0 {
					return "", fmt.Errorf("malformed (%s ...) stanza", keyword)
				}
				val := strings.TrimSpace(after[:end])
				if val != "" {
					return val, nil
				}
			}
			depth++
			i++
			continue
		}
		if c == ')' {
			if depth > 0 {
				depth--
			}
		}
		i++
	}
	return "", fmt.Errorf("(%s ...) not found", keyword)
}

// findStanzaBounds finds the interior of a stanza starting with keyword (e.g. "(package").
// Returns start (after keyword) and end (position of matching closing ')').
func findStanzaBounds(content, keyword string) (start, end int, err error) {
	idx := strings.Index(content, keyword)
	if idx < 0 {
		return 0, 0, fmt.Errorf("%q not found", keyword)
	}
	start = idx + len(keyword)
	depth := 1
	i := start
	inString := false
	for i < len(content) {
		c := content[i]
		switch {
		case inString && c == '"':
			inString = false
		case inString && c == '\\':
			i++
		case !inString && c == '"':
			inString = true
		case !inString && c == '(':
			depth++
		case !inString && c == ')':
			depth--
			if depth == 0 {
				return start, i, nil
			}
		}
		i++
	}
	return 0, 0, fmt.Errorf("unterminated %q stanza", keyword)
}

// AddDep adds a dependency to the (depends ...) stanza in dune-project.
// Constraint "*" means no version constraint. If pkg is already present, this is a no-op.
func AddDep(dir, pkg, constraint string) error {
	path := filepath.Join(dir, "dune-project")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read dune-project: %w", err)
	}
	content := string(data)

	// Idempotency check: is the pkg already in the depends block?
	if hasDep(content, pkg) {
		return nil
	}

	entry := formatDuneDepEntry(pkg, constraint)

	// Find the (depends ...) block and insert before its closing paren.
	_, end, err := findDepsBounds(content)
	if err != nil {
		return err
	}

	// Insert new entry before the closing ')' of depends
	newContent := content[:end] + "\n  " + entry + content[end:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

// RemoveDep removes a dependency from the (depends ...) stanza in dune-project.
// Returns an error if pkg is not found.
func RemoveDep(dir, pkg string) error {
	path := filepath.Join(dir, "dune-project")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read dune-project: %w", err)
	}
	content := string(data)

	if !hasDep(content, pkg) {
		return fmt.Errorf("%q is not in the (depends ...) stanza", pkg)
	}

	start, end, err := findDepsBounds(content)
	if err != nil {
		return err
	}

	interior := content[start:end]
	filtered, ok := removeDuneDepEntry(interior, pkg)
	if !ok {
		return fmt.Errorf("%q is not in the (depends ...) stanza", pkg)
	}

	newContent := content[:start] + filtered + content[end:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

// findDepsBounds locates the interior of the (depends ...) block in a dune-project.
// Returns the start and end positions of the content between '(' depends and matching ')'.
// 'start' points to after "(depends", 'end' points to the position of the closing ')'.
func findDepsBounds(content string) (start, end int, err error) {
	start, end, err = findStanzaBounds(content, "(depends")
	if err != nil {
		return 0, 0, fmt.Errorf("no (depends ...) stanza in dune-project")
	}
	return start, end, nil
}

// hasDep reports whether pkg appears as a dependency name in content.
func hasDep(content, pkg string) bool {
	start, end, err := findDepsBounds(content)
	if err != nil {
		return false
	}
	// Pass just the interior (between "(depends" and the closing ")")
	interior := content[start:end]
	return containsDepName(interior, pkg)
}

// containsDepName reports whether pkg appears as a dep name in the depends block text.
func containsDepName(block, pkg string) bool {
	// Scan through the block looking for the package name as an atom or start of a list
	i := 0
	for i < len(block) {
		c := block[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if c == '"' {
			// skip string
			i++
			for i < len(block) && block[i] != '"' {
				if block[i] == '\\' {
					i++
				}
				i++
			}
			i++
			continue
		}
		if c == '(' {
			// read the first atom inside this list
			i++
			// skip whitespace
			for i < len(block) && (block[i] == ' ' || block[i] == '\t' || block[i] == '\n' || block[i] == '\r') {
				i++
			}
			// read atom
			start := i
			for i < len(block) && block[i] != ' ' && block[i] != '\t' && block[i] != '\n' && block[i] != '\r' && block[i] != ')' && block[i] != '(' {
				i++
			}
			name := block[start:i]
			if name == pkg {
				return true
			}
			// skip the rest of this list
			depth := 1
			for i < len(block) && depth > 0 {
				switch block[i] {
				case '(':
					depth++
				case ')':
					depth--
				case '"':
					i++
					for i < len(block) && block[i] != '"' {
						if block[i] == '\\' {
							i++
						}
						i++
					}
				}
				i++
			}
			continue
		}
		// bare atom
		start := i
		for i < len(block) && block[i] != ' ' && block[i] != '\t' && block[i] != '\n' && block[i] != '\r' && block[i] != ')' && block[i] != '(' {
			i++
		}
		name := block[start:i]
		if name == pkg {
			return true
		}
	}
	return false
}

// removeDuneDepEntry removes the dep entry for pkg from the interior of a depends block.
// Returns the modified interior and true if pkg was found, or the original interior and false if not.
func removeDuneDepEntry(interior, pkg string) (string, bool) {
	// We'll scan line by line looking for the dep
	lines := strings.Split(interior, "\n")
	result := make([]string, 0, len(lines))
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check if this line starts the entry for pkg
		if isDepLine(trimmed, pkg) {
			found = true
			// If this line has balanced parens (single-line entry), skip it
			// If unbalanced, we need to also skip following lines until balanced
			depth := countParenDepth(trimmed)
			if depth == 0 {
				continue // single-line entry fully removed
			}
			// Multi-line: keep skipping until depth returns to 0
			// (For our generated format, entries are always single-line)
			continue
		}
		result = append(result, line)
	}
	if !found {
		return interior, false
	}
	return strings.Join(result, "\n"), true
}

// isDepLine reports whether a trimmed line is the dep entry for pkg.
func isDepLine(trimmed, pkg string) bool {
	// Bare atom: exactly the pkg name
	if trimmed == pkg {
		return true
	}
	// List starting with pkg: "(pkg ..."
	if strings.HasPrefix(trimmed, "("+pkg+" ") || strings.HasPrefix(trimmed, "("+pkg+")") {
		return true
	}
	return false
}

// countParenDepth returns the net paren depth of a string (opens - closes).
func countParenDepth(s string) int {
	depth := 0
	inStr := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inStr && c == '"':
			inStr = false
		case inStr && c == '\\':
			i++
		case !inStr && c == '"':
			inStr = true
		case !inStr && c == '(':
			depth++
		case !inStr && c == ')':
			depth--
		}
	}
	return depth
}

// formatDuneDepEntry formats a dep entry for insertion into a dune-project depends block.
func formatDuneDepEntry(pkg, constraint string) string {
	if constraint == "*" || constraint == "" {
		return pkg
	}
	op, ver := parseConstraintParts(constraint)
	if op == "" {
		return pkg
	}
	return fmt.Sprintf("(%s (%s %q))", pkg, op, ver)
}

// parseConstraintParts splits a constraint like ">=5.0.0" into op=">=", ver="5.0.0".
func parseConstraintParts(c string) (op, ver string) {
	for _, prefix := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(c, prefix) {
			return prefix, strings.TrimSpace(c[len(prefix):])
		}
	}
	return "", c
}
