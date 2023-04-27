package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
)

func GetConfig(profileName string, regionName string) aws.Config {
	// ref : https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/
	//       https://zenn.dev/kz23szk/articles/f3e8fc167fdeeb
	var optFunctions = make([]func(*config.LoadOptions) error, 0)

	if regionName != "" {
		optFunctions = append(optFunctions, config.WithRegion(regionName))
	} else {
		// try to get region name from IMDS
		if _, exists := os.LookupEnv("AWS_EXECUTION_ENV"); exists {
			optFunctions = append(optFunctions, config.WithEC2IMDSRegion())
		}
	}
	if profileName != "" {
		optFunctions = append(optFunctions, config.WithSharedConfigProfile(profileName))
	}
	optFunctions = append(optFunctions, config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
		options.TokenProvider = func() (string, error) {
			return stscreds.StdinTokenProvider()
		}
	}))

	cfg, err := config.LoadDefaultConfig(context.Background(), optFunctions...)
	if err != nil {
		panic(fmt.Sprintf("aws configuration error, %v", err.Error()))
	}
	return cfg
}
