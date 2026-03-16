package cmd

import "path/filepath"

// relPath returns the path of target relative to base.
// If filepath.Rel fails (which should not happen in practice), it falls back to target.
func relPath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}
