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
	"github.com/aws/smithy-go"
)

type EC2API interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)

	GetPasswordData(ctx context.Context, params *ec2.GetPasswordDataInput, optFns ...func(*ec2.Options)) (*ec2.GetPasswordDataOutput, error)
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
	result, err := api.DescribeInstances(ctx, input)
	if err != nil {
		return "", err
	}
	for _, r := range result.Reservations {
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
