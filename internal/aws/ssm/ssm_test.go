package ssm

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type MockAPI struct {
	DescribeInstanceInformationOutput *ssm.DescribeInstanceInformationOutput
	StartSessionOutput                *ssm.StartSessionOutput
	TerminateSessionOutput            *ssm.TerminateSessionOutput
	Error                             error
}

func (m *MockAPI) DescribeInstanceInformation(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
	return m.DescribeInstanceInformationOutput, m.Error
}

func (m *MockAPI) StartSession(ctx context.Context, params *ssm.StartSessionInput, optFns ...func(*ssm.Options)) (*ssm.StartSessionOutput, error) {
	return m.StartSessionOutput, m.Error
}

func (m *MockAPI) TerminateSession(ctx context.Context, params *ssm.TerminateSessionInput, optFns ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error) {
	return m.TerminateSessionOutput, m.Error
}

func Test_IsInstanceOnline(t *testing.T) {
	var instanceId = "i-1234567890"

	// when instance is not online (Inactive)
	var mock = &MockAPI{
		DescribeInstanceInformationOutput: &ssm.DescribeInstanceInformationOutput{
			InstanceInformationList: []types.InstanceInformation{{PingStatus: types.PingStatusInactive}},
		},
		Error: nil,
	}
	var result, err = IsInstanceOnline(mock, context.TODO(), instanceId)
	if err == nil {
		t.Error("Failed to get instance status")
	}
	if result {
		t.Error("Instance status is Inactive")
	}

	// when instance is not online (ConnectionLost)
	mock = &MockAPI{
		DescribeInstanceInformationOutput: &ssm.DescribeInstanceInformationOutput{
			InstanceInformationList: []types.InstanceInformation{{PingStatus: types.PingStatusConnectionLost}},
		},
		Error: nil,
	}
	result, err = IsInstanceOnline(mock, context.TODO(), instanceId)
	if err == nil {
		t.Error("Failed to get instance status")
	}
	if result {
		t.Error("Instance status is ConnectionLost")
	}

	// when instance is online
	mock = &MockAPI{
		DescribeInstanceInformationOutput: &ssm.DescribeInstanceInformationOutput{
			InstanceInformationList: []types.InstanceInformation{{PingStatus: types.PingStatusOnline}},
		},
		Error: nil,
	}
	result, err = IsInstanceOnline(mock, context.TODO(), instanceId)
	if err != nil {
		t.Error("Failed to get instance status")
	}
	if !result {
		t.Error("Instance status is Online")
	}
}
