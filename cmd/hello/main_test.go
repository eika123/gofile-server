package main

import (
	"path/filepath"
	"testing"
)

func TestInTrustedRoot(t *testing.T) {
	root, _ := filepath.Abs("./test_root")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"Inside root", filepath.Join(root, "file.txt"), false},
		{"Inside root subdirectory", filepath.Join(root, "subdir", "file.txt"), false},
		{"Root itself", root, false},
		{"Traversing up to root", filepath.Join(root, "..", "test_root", "file.txt"), false},
		{"Parent of root", filepath.Join(root, ".."), true},
		{"Outside sibling", filepath.Join(root, "..", "sibling"), true},
		{"Absolute outside", "/etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := inTrustedRoot(tt.path, root)
			if (err != nil) != tt.wantErr {
				t.Errorf("inTrustedRoot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}	