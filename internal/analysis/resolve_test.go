package analysis

import (
	"testing"
)

// TestResolveFilePaths_RelativeToAbsolute verifies that resolveFilePaths
// converts relative paths to absolute using the given project root, while
// leaving already-absolute paths unchanged.
func TestResolveFilePaths_RelativeToAbsolute(t *testing.T) {
	root := "/home/user/project"

	cases := []struct {
		in   string
		want string
	}{
		{"internal/pkg/file.go", "/home/user/project/internal/pkg/file.go"},
		{"/abs/path/other.go", "/abs/path/other.go"},
		{"README.md", "/home/user/project/README.md"},
	}

	for _, tc := range cases {
		got := resolveFilePaths([]string{tc.in}, root)
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("resolveFilePaths(%q, %q) = %v, want [%q]", tc.in, root, got, tc.want)
		}
	}
}

// TestResolveFilePaths_EmptyRoot verifies that when projectRoot is empty,
// paths are returned unchanged (graceful degradation).
func TestResolveFilePaths_EmptyRoot(t *testing.T) {
	files := []string{"relative/path.go", "/abs/path.go"}
	got := resolveFilePaths(files, "")
	for i, f := range files {
		if got[i] != f {
			t.Errorf("resolveFilePaths with empty root: index %d got %q, want %q", i, got[i], f)
		}
	}
}
