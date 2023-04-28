package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_validatePemFile(t *testing.T) {
	var tempDir = t.TempDir()
	// Raise error when argument is ""
	err := validatePemFile("")
	if err == nil {
		t.Error("Argument must be specified")
	}
	// Raise error when .pem file not exists
	err = validatePemFile(filepath.Join(tempDir, "non-existent-test.pem"))
	if err == nil {
		t.Error(".pem file must be exist")
	}
	// Success when .pem file exists
	fileName := filepath.Join(tempDir, "exists-test.pem")
	f, _ := os.Create(fileName)
	t.Logf("%v created...", fileName)
	defer f.Close()
	err = validatePemFile(fileName)
	if err != nil {
		t.Error(".pem file exists")
	}
}

func Test_validatePort(t *testing.T) {
	// Valid port is 1 to 65535
	cases := []struct {
		Port  int
		Valid bool
	}{
		{0, false},
		{1, true},
		{3389, true},
		{65535, true},
		{65536, false},
	}
	for _, c := range cases {
		var err = validatePort(c.Port)
		if c.Valid {
			if err != nil {
				t.Errorf("Port %v is valid", c.Port)
			}
		} else {
			if err == nil {
				t.Errorf("Port %v is invalid", c.Port)
			}
		}
	}
}
