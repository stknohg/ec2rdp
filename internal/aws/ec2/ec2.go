package ec2

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

type EC2API interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)

	DescribeInstanceConnectEndpoints(ctx context.Context, params *ec2.DescribeInstanceConnectEndpointsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceConnectEndpointsOutput, error)

	GetPasswordData(ctx context.Context, params *ec2.GetPasswordDataInput, optFns ...func(*ec2.Options)) (*ec2.GetPasswordDataOutput, error)
}

type InstanceMetadataForEICE struct {
	State            types.InstanceState
	PrivateIpAddress string
	VpcId            string
}

type EICEndpointMetadata struct {
	EndpointId  string
	DnsName     string
	FipsDnsName string
}

func NewAPI(cfg aws.Config) EC2API {
	return ec2.NewFromConfig(cfg)
}

func IsInstanceExist(api EC2API, ctx context.Context, instanceId string) (bool, error) {
	input := &ec2.DescribeInstancesInput{InstanceIds: []string{instanceId}}
	_, err := api.DescribeInstances(ctx, input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "InvalidInstanceID.Malformed" {
				return false, fmt.Errorf("instance %v not found", instanceId)
			}
		}
		return false, err
	}
	return true, nil
}

func GetPublicHostName(api EC2API, ctx context.Context, instanceId string) (string, error) {
	input := &ec2.DescribeInstancesInput{InstanceIds: []string{instanceId}}
	output, err := api.DescribeInstances(ctx, input)
	if err != nil {
		return "", err
	}
	for _, r := range output.Reservations {
		for _, i := range r.Instances {
			hostname := i.PublicDnsName
			if hostname != nil && *hostname != "" {
				return *hostname, nil
			}
			publicIP := i.PublicIpAddress
			if publicIP != nil && *publicIP != "" {
				return *publicIP, nil
			}
		}
	}
	return "", errors.New("failed to find public hostname")
}

func GetInstanceMetadataForEICE(api EC2API, ctx context.Context, instanceId string) (*InstanceMetadataForEICE, error) {
	input := &ec2.DescribeInstancesInput{InstanceIds: []string{instanceId}}
	output, err := api.DescribeInstances(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(output.Reservations) == 0 {
		return nil, fmt.Errorf("failed to find instance reservation")
	}
	if len(output.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("failed to find instance")
	}
	return &InstanceMetadataForEICE{
		State:            *output.Reservations[0].Instances[0].State,
		PrivateIpAddress: *output.Reservations[0].Instances[0].PrivateIpAddress,
		VpcId:            *output.Reservations[0].Instances[0].VpcId,
	}, nil
}

func fetchEICEndpoint(api EC2API, ctx context.Context, input *ec2.DescribeInstanceConnectEndpointsInput) (*EICEndpointMetadata, error) {
	output, err := api.DescribeInstanceConnectEndpoints(ctx, input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "InvalidInstanceConnectEndpointId.NotFound" {
				return nil, fmt.Errorf("EC2 Instance Connect Endpoint ID is invalid")
			}
		}
		return nil, err
	}
	if len(output.InstanceConnectEndpoints) == 0 {
		return nil, fmt.Errorf("EC2 Instance Connect Endpoint is not found")
	}
	result := EICEndpointMetadata{
		EndpointId: *output.InstanceConnectEndpoints[0].InstanceConnectEndpointId,
		DnsName:    *output.InstanceConnectEndpoints[0].DnsName,
	}
	if output.InstanceConnectEndpoints[0].FipsDnsName != nil {
		result.FipsDnsName = *output.InstanceConnectEndpoints[0].FipsDnsName
	}
	return &result, nil
}

func FetchEICEndpointById(api EC2API, ctx context.Context, endpointId string) (*EICEndpointMetadata, error) {
	input := &ec2.DescribeInstanceConnectEndpointsInput{}
	if endpointId != "" {
		input.InstanceConnectEndpointIds = []string{endpointId}
	}
	filters := []types.Filter{}
	filters = append(filters, types.Filter{Name: aws.String("state"), Values: []string{"create-complete"}})
	input.Filters = filters
	return fetchEICEndpoint(api, ctx, input)
}

func FetchEICEndpointByVpc(api EC2API, ctx context.Context, vpcId string) (*EICEndpointMetadata, error) {
	input := &ec2.DescribeInstanceConnectEndpointsInput{}
	filters := []types.Filter{}
	filters = append(filters, types.Filter{Name: aws.String("state"), Values: []string{"create-complete"}})
	filters = append(filters, types.Filter{Name: aws.String("vpc-id"), Values: []string{vpcId}})
	input.Filters = filters
	return fetchEICEndpoint(api, ctx, input)
}

func GetAdministratorPassword(api EC2API, ctx context.Context, instanceId string, pemFilePath string) (string, error) {
	input := &ec2.GetPasswordDataInput{InstanceId: &instanceId}
	result, err := api.GetPasswordData(ctx, input)
	if err != nil {
		return "", err
	}

	// decrypt password
	passowrd, err := decodePassword(*result.PasswordData, pemFilePath)
	if err != nil {
		return "", err
	}
	return passowrd, nil
}

func decodePassword(passwordData string, pemFilePath string) (string, error) {
	if passwordData == "" {
		return "", nil
	}

	// ref : https://github.com/tomrittervg/decrypt-windows-ec2-passwd/blob/master/decrypt-windows-ec2-passwd.go

	// base64 decode
	encPassword, err := base64.StdEncoding.DecodeString(passwordData)
	if err != nil {
		return "", err
	}

	// get private key
	pemKey, err := getPemKey(pemFilePath)
	if err != nil {
		return "", err
	}

	// decrypt password
	binPassword, err := rsa.DecryptPKCS1v15(nil, pemKey, encPassword)
	if err != nil {
		return "", err
	}
	return string(binPassword), nil
}

func getPemKey(pemFilePath string) (*rsa.PrivateKey, error) {
	rawBytes, err := os.ReadFile(pemFilePath)
	if err != nil {
		return &rsa.PrivateKey{}, err
	}
	block, _ := pem.Decode(rawBytes)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return &rsa.PrivateKey{}, err
	}
	if err := key.Validate(); err != nil {
		return &rsa.PrivateKey{}, err
	}
	return key, nil
}
