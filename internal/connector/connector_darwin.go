//go:build darwin

package connector

import (
	"fmt"
	"os/exec"
)

func (f *DefaultConnector) PreConnect() error {
	// do nothing
	return nil
}

func (f *DefaultConnector) Connect() error {
	// start Parallels Client
	fmt.Printf("Connect to %v:%v\n", f.HostName, f.Port)
	rasUrl = fmt.Sprintf("tuxclient:///?Command=LaunchApp&ConnType=2&Server=%v&Backup=&Port=%v&UserName=%v&Password=%v", f.HostName, f.Port, f.UserName, f.PlainPassword)
	cmd := exec.Command("open", rasUrl)
	if f.WaitFor {
		cmd.Run()
	} else {
		cmd.Start()
	}
	return nil
}

func (f *DefaultConnector) PostConnect() error {
	// do nothing
	return nil
}
