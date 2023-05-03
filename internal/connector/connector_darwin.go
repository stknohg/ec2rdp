//go:build darwin

package connector

import (
	"fmt"
	"os"
	"os/exec"
)

func (f *DefaultConnector) IsInstalled() (bool, error) {
	_, err := os.Stat("/Applications/Parallels Client.app")
	if err != nil {
		return false, fmt.Errorf("%v is not installed", "Parallels Client")
	}
	return true, nil
}

func (f *DefaultConnector) PreConnect() error {
	// do nothing
	return nil
}

func (f *DefaultConnector) Connect() error {
	// start Parallels Client
	fmt.Printf("Connect to %v:%v\n", f.HostName, f.Port)
	var rasUrl = fmt.Sprintf("tuxclient:///?Command=LaunchApp&ConnType=2&Server=%v&Backup=&Port=%v&LoginEx=%v&Password=%v", f.HostName, f.Port, f.UserName, f.PlainPassword)
	cmd := exec.Command("open", rasUrl)
	cmd.Start()
	if f.WaitFor {
		// To prevent password appearing from arguments, wait for the .app process.
		cmd := exec.Command("open", "--wait-apps", "/Applications/Parallels Client.app")
		cmd.Run()
	}
	return nil
}

func (f *DefaultConnector) PostConnect() error {
	// do nothing
	return nil
}
