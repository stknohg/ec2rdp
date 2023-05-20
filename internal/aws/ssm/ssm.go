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

type SSMAPI interface {
	DescribeInstanceInformation(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error)

	StartSession(ctx context.Context, params *ssm.StartSessionInput, optFns ...func(*ssm.Options)) (*ssm.StartSessionOutput, error)

	TerminateSession(ctx context.Context, params *ssm.TerminateSessionInput, optFns ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error)
}

type StartSSMSessionPluginResult struct {
	API       SSMAPI
	SessionId string
	ProcessId int
}

type sessionManagerPluginParameter struct {
	Target     string
	Parameters map[string][]string
}

func NewAPI(cfg aws.Config) SSMAPI {
	return ssm.NewFromConfig(cfg)
}

func IsInstanceOnline(api SSMAPI, ctx context.Context, instanceId string) (bool, error) {
	input := &ssm.DescribeInstanceInformationInput{
		InstanceInformationFilterList: []types.InstanceInformationFilter{
			{Key: "InstanceIds", ValueSet: []string{instanceId}},
		},
	}
	result, err := api.DescribeInstanceInformation(ctx, input)
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

func StartSSMSessionPortForward(api SSMAPI, ctx context.Context, instanceId string, port int, localPort int, reason string, region string, profile string) (*StartSSMSessionPluginResult, error) {
	return StartSSMSessionWithPlugin(
		api,
		ctx,
		instanceId,
		"AWS-StartPortForwardingSession",
		map[string][]string{"portNumber": {strconv.Itoa(port)}, "localPortNumber": {strconv.Itoa(localPort)}},
		reason,
		region,
		profile)
}

func StartSSMSessionWithPlugin(api SSMAPI, ctx context.Context, target string, documentName string, parameters map[string][]string, reason string, region string, profile string) (*StartSSMSessionPluginResult, error) {
	if target == "" {
		return &StartSSMSessionPluginResult{}, fmt.Errorf("no target specified")
	}
	if region == "" {
		return &StartSSMSessionPluginResult{}, fmt.Errorf("no region name specified")
	}

	// start session
	input := &ssm.StartSessionInput{
		Target:       &target,
		DocumentName: &documentName,
		Parameters:   parameters,
		Reason:       &reason,
	}
	result, err := api.StartSession(ctx, input)
	if err != nil {
		return &StartSSMSessionPluginResult{}, err
	}

	// start session manager plugin
	// arg1
	sessionJson, _ := json.Marshal(result)
	arg1 := string(sessionJson)
	// arg1
	arg2 := region
	// arg3
	arg3 := "StartSession"
	// arg4
	arg4 := profile
	// arg5
	pluginParameter := &sessionManagerPluginParameter{Target: target, Parameters: input.Parameters}
	parameterJson, _ := json.Marshal(pluginParameter)
	arg5 := string(parameterJson)
	// arg6
	arg6 := fmt.Sprintf("https://ssm.%v.amazonaws.com", region)
	// start process
	cmd := exec.Command("session-manager-plugin", arg1, arg2, arg3, arg4, arg5, arg6)
	err = cmd.Start()
	return &StartSSMSessionPluginResult{API: api, SessionId: *result.SessionId, ProcessId: cmd.Process.Pid}, err
}

func TerminateSSMSession(api SSMAPI, ctx context.Context, sessionId string) error {
	// start session
	input := &ssm.TerminateSessionInput{
		SessionId: &sessionId,
	}
	_, err := api.TerminateSession(ctx, input)
	if err != nil {
		return err
	}
	return nil
}
