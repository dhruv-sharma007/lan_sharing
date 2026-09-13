package transfer

import (
	"path/filepath"
	"testing"
)

func TestValidateAndJoinPath(t *testing.T) {
	baseDir := filepath.Join("C:", "LanShare", "Received")

	tests := []struct {
		name        string
		relPath     string
		expectError bool
	}{
		{"Valid simple file", "photo.jpg", false},
		{"Valid nested file", "photos/vacation/photo.jpg", false},
		{"Valid nested file with backslashes", "photos\\vacation\\photo.jpg", false},
		{"Empty path", "", true},
		{"Absolute path unix", "/etc/passwd", true},
		{"Absolute path windows", "C:\\Windows\\System32\\cmd.exe", true},
		{"Relative escaping unix", "../photo.jpg", true},
		{"Relative escaping multiple unix", "../../photo.jpg", true},
		{"Relative escaping windows", "..\\photo.jpg", true},
		{"Internal relative valid", "photos/../photo.jpg", false},
		{"Sneaky escaping", "photos/../../photo.jpg", true},
		{"Root path", "/", true},
		{"Root path windows", "\\", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateAndJoinPath(baseDir, tt.relPath)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}
