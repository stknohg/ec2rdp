package ec2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

type MockAPI struct {
	DescribeInstancesOutput *ec2.DescribeInstancesOutput
	GetPasswordDataOutput   *ec2.GetPasswordDataOutput
	Error                   error
}

func (m *MockAPI) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.DescribeInstancesOutput, m.Error
}

func (m *MockAPI) GetPasswordData(ctx context.Context, params *ec2.GetPasswordDataInput, optFns ...func(*ec2.Options)) (*ec2.GetPasswordDataOutput, error) {
	return m.GetPasswordDataOutput, m.Error
}

func Test_IsInstanceExist(t *testing.T) {
	// when instance exists
	var instanceId = "i-1234567890"
	var mock = &MockAPI{
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{{Instances: []types.Instance{{InstanceId: &instanceId}}}},
		},
		Error: nil,
	}
	if exists, _ := IsInstanceExist(mock, instanceId); !exists {
		t.Error("Instance exists")
	}

	// when instance not exists
	mock = &MockAPI{
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{},
		Error:                   &smithy.GenericAPIError{Code: "InvalidInstanceID.Malformed"},
	}
	if exists, _ := IsInstanceExist(mock, instanceId); exists {
		t.Error("Instance exists")
	}
}

func Test_GetPublicHostName(t *testing.T) {
	var instanceId = "i-1234567890"
	var publicDNSName = "public.example.com"
	var publicIP = "1.2.3.4"

	// when public DNS name exists
	var mock = &MockAPI{
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{{Instances: []types.Instance{{InstanceId: &instanceId, PublicDnsName: &publicDNSName, PublicIpAddress: &publicIP}}}},
		},
		Error: nil,
	}
	var result, err = GetPublicHostName(mock, instanceId)
	if err != nil {
		t.Error("Failed to get public DNS name")
	}
	if result != publicDNSName {
		t.Error("Invalid public DNS name")
	}

	// when public IP exists
	mock = &MockAPI{
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{{Instances: []types.Instance{{InstanceId: &instanceId, PublicIpAddress: &publicIP}}}},
		},
		Error: nil,
	}
	result, err = GetPublicHostName(mock, instanceId)
	if err != nil {
		t.Error("Failed to get public IP address")
	}
	if result != publicIP {
		t.Error("Invalid public IP address")
	}

	// when instance stopped
	mock = &MockAPI{
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{{Instances: []types.Instance{{InstanceId: &instanceId}}}},
		},
		Error: nil,
	}
	_, err = GetPublicHostName(mock, instanceId)
	if err == nil {
		t.Error("Failed to get public DNS name or IP address")
	}
	if err.Error() != "failed to find public hostname" {
		t.Error("Invalid error message")
	}
}

func Test_GetAdministratorPassword(t *testing.T) {
	// when password data not found
	var instanceId = "i-1234567890"
	var encodedPassword = "" // PassworData is empty
	var expedtedPassword = ""
	var mock = &MockAPI{
		GetPasswordDataOutput: &ec2.GetPasswordDataOutput{
			PasswordData: &encodedPassword,
		},
		Error: nil,
	}
	var result, err = GetAdministratorPassword(mock, instanceId, "./testdata/test.pem")
	if err != nil {
		t.Error("Failed to get PasswordData")
	}
	if result != expedtedPassword {
		t.Error("PasswordData is empty")
	}

	// when password data exists
	encodedPassword = "ilVJituy4wak95QClqnC/FcUbQWTZHXaCNR5yMvxL24TDeWaoSlnPxS5eIX07tEAZHmgqINGc1cD5tKMEHgO47+lt1p7vvB5mXYDdrwVAuSA5K8tg7BIA7umYlgVIocNTzUJHEmr10Lx/Vlb3g1AEE9Rl1fnk7FYCl6kBkwpejcCtqLZclt2wt62GkGR5KekHAsw3Fiy4x9uMUkgfjwH7FjFld+FzZUJ1RNrCC7H6dvnk1WIbgnQetwecAFq56heimDD7BKncsAu5R0gOMEGB88KLzjEPJi5c6T73e/W3jvD7us4evRUFIM7tcaQ8RBmBa7eDYmXFIEcmfGRm38Trg=="
	expedtedPassword = "4Hio.kdu40ajlj%p7ZfINkkR5uU6e-zY"
	mock = &MockAPI{
		GetPasswordDataOutput: &ec2.GetPasswordDataOutput{
			PasswordData: &encodedPassword,
		},
		Error: nil,
	}
	result, err = GetAdministratorPassword(mock, instanceId, "./testdata/test.pem")
	if err != nil {
		t.Error("Failed to get PasswordData")
	}
	if result != expedtedPassword {
		t.Error("Dedode password is wrong")
	}
}