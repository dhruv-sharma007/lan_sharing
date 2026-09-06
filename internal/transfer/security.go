package transfer

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidPath = errors.New("invalid or malicious relative path")
)

// ValidateAndJoinPath securely joins a base directory with a relative path.
// It ensures that the resulting path is contained within the base directory
// and prevents directory traversal attacks (e.g., ../ or absolute paths).
func ValidateAndJoinPath(baseDir, relPath string) (string, error) {
	if relPath == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	// Clean the relative path to resolve any ../ internally first
	cleanRel := filepath.Clean(relPath)

	// In Go, Clean might return absolute path if relPath was absolute.
	// But it might also return a volume name on Windows.
	// On Windows, / or \ might not be considered absolute by IsAbs (relative to current drive).
	if filepath.IsAbs(cleanRel) || filepath.VolumeName(cleanRel) != "" || strings.HasPrefix(cleanRel, "/") || strings.HasPrefix(cleanRel, "\\") {
		return "", fmt.Errorf("%w: absolute or volume paths are not allowed", ErrInvalidPath)
	}

	// Double check that it doesn't start with ..
	if cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) || strings.HasPrefix(cleanRel, "../") {
		return "", fmt.Errorf("%w: path escapes base directory", ErrInvalidPath)
	}

	// Abs the base dir to make comparison robust
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to absolute base dir: %w", err)
	}

	finalPath := filepath.Join(absBase, cleanRel)

	// Extra safety check: ensuring finalPath starts with absBase
	// (filepath.Join handles separators properly, but this is defense in depth)
	if !strings.HasPrefix(finalPath, absBase) {
		return "", fmt.Errorf("%w: final path escapes base directory", ErrInvalidPath)
	}

	return finalPath, nil
}
