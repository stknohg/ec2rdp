package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type StartSSMSessionPluginResult struct {
	Config    aws.Config
	SessionId string
	ProcessId int
}

type sessionManagerPluginParameter struct {
	Target     string
	Parameters map[string][]string
}

func IsInstanceOnline(cfg aws.Config, instanceId string) (bool, error) {
	client := ssm.NewFromConfig(cfg)

	input := &ssm.DescribeInstanceInformationInput{
		InstanceInformationFilterList: []types.InstanceInformationFilter{
			{Key: "InstanceIds", ValueSet: []string{instanceId}},
		},
	}
	result, err := client.DescribeInstanceInformation(context.TODO(), input)
	if err != nil {
		return false, err
	}
	if len(result.InstanceInformationList) == 0 {
		return false, fmt.Errorf("instance %v is not online", instanceId)
	}
	status := result.InstanceInformationList[0].PingStatus
	if status == types.PingStatusOnline {
		return true, nil
	}
	return false, fmt.Errorf("instance %v is not online. (SSM PingStatus : %v)", instanceId, status)
}

func StartSSMSessionWithPlugin(cfg aws.Config, instanceId string, port int, localPort int, profileName string) (*StartSSMSessionPluginResult, error) {
	client := ssm.NewFromConfig(cfg)

	// start session
	input := &ssm.StartSessionInput{
		Target:       &instanceId,
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters:   map[string][]string{"portNumber": {strconv.Itoa(port)}, "localPortNumber": {strconv.Itoa(localPort)}},
		Reason:       aws.String("ec2rdp ssm"),
	}
	result, err := client.StartSession(context.TODO(), input)
	if err != nil {
		return &StartSSMSessionPluginResult{}, err
	}

	// start session manager plugin
	var ssmRegion = cfg.Region
	// arg1
	sessionJson, _ := json.Marshal(result)
	arg1 := string(sessionJson)
	// arg1
	arg2 := ssmRegion
	// arg3
	arg3 := "StartSession"
	// arg4
	arg4 := profileName
	// arg5
	pluginParameter := &sessionManagerPluginParameter{Target: instanceId, Parameters: input.Parameters}
	parameterJson, _ := json.Marshal(pluginParameter)
	arg5 := string(parameterJson)
	// arg6
	arg6 := fmt.Sprintf("https://ssm.%v.amazonaws.com", ssmRegion)
	// start process
	cmd := exec.Command("session-manager-plugin", arg1, arg2, arg3, arg4, arg5, arg6)
	err = cmd.Start()
	return &StartSSMSessionPluginResult{Config: cfg, SessionId: *result.SessionId, ProcessId: cmd.Process.Pid}, err
}

func TerminateSSMSession(cfg aws.Config, sessionId string) error {
	client := ssm.NewFromConfig(cfg)

	// start session
	input := &ssm.TerminateSessionInput{
		SessionId: &sessionId,
	}
	_, err := client.TerminateSession(context.TODO(), input)
	if err != nil {
		return err
	}
	return nil
}
