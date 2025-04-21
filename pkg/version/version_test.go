package version

import (
	"runtime/debug"
	"testing"
)

func TestVersion(t *testing.T) {
	// Save original and restore after test
	originalReadBuildInfo := readBuildInfo
	defer func() { readBuildInfo = originalReadBuildInfo }()

	// Test when build info is available
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		info := &debug.BuildInfo{
			Main: debug.Module{Version: "1.2.3"},
		}
		return info, true
	}
	if v := Version(); v != "1.2.3" {
		t.Errorf("Expected '1.2.3', got %q", v)
	}

	// Test when build info is not available
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return nil, false
	}
	if v := Version(); v != "unknown" {
		t.Errorf("Expected 'unknown', got %q", v)
	}
}
