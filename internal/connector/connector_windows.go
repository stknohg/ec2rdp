//go:build windows

package connector

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/danieljoos/wincred"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func (f *DefaultConnector) PreConnect() error {
	fmt.Printf("Save credential TERMSRV/%v to Credential Manager\n", f.HostName)
	//cmd := exec.Command("cmdkey", fmt.Sprintf("/generic:TERMSRV/%v", f.HostName), fmt.Sprintf("/user:%v", f.UserName), fmt.Sprintf("/pass:%v", f.PlainPassword))
	//cmd.Run()

	// get password blob
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	blob, _, err := transform.Bytes(encoder, []byte(f.PlainPassword))
	if err != nil {
		return err
	}
	// save credential
	cred := wincred.NewGenericCredential(fmt.Sprintf("TERMSRV/%v", f.HostName))
	cred.Persist = wincred.PersistEnterprise
	cred.UserName = f.UserName
	cred.CredentialBlob = blob
	return cred.Write()
}

func (f *DefaultConnector) Connect() error {
	// invoke mstsc
	fmt.Printf("Connect to %v:%v\n", f.HostName, f.Port)
	cmd := exec.Command("mstsc", fmt.Sprintf("/v:%v:%v", f.HostName, f.Port), "/f")
	if f.WaitFor {
		cmd.Run()
	} else {
		cmd.Start()
		// wait minimum time for RDP client to use credential.
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (f *DefaultConnector) PostConnect() error {
	fmt.Printf("Delete credential TERMSRV/%v from Credential Manager\n", f.HostName)
	//cmd := exec.Command("cmdkey", fmt.Sprintf("/delete:TERMSRV/%v", f.HostName))
	//cmd.Run()

	// delete credential
	cred, err := wincred.GetGenericCredential(fmt.Sprintf("TERMSRV/%v", f.HostName))
	if err != nil {
		return err
	}
	return cred.Delete()
}
