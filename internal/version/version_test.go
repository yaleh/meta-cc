package version

import (
	"regexp"
	"testing"
)

var semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func TestVersionIsSemver(t *testing.T) {
	if !semverPattern.MatchString(Version) {
		t.Fatalf("version.Version = %q is not a valid X.Y.Z release version (from internal/version/release.json)", Version)
	}
}

func TestServerNameIsSet(t *testing.T) {
	if ServerName == "" {
		t.Fatal("version.ServerName must not be empty")
	}
}
