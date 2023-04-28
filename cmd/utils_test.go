package cmd

import "testing"

func Test_isPortOpen(t *testing.T) {
	// Fail when invalid port is specified
	if isPortOpen("localhost", 0) {
		t.Error("Valid port range is 1 to 65535")
	}
	if isPortOpen("localhost", 65536) {
		t.Error("Valid port range is 1 to 65535")
	}
	// Fail when non-existent hostname
	if isPortOpen("non-existent.local", 80) {
		t.Error("Must fail when non-existent hostname is specified")
	}
}
