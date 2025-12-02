module github.com/stknohg/ec2rdp

go 1.25.0

require (
	github.com/aws/aws-sdk-go-v2 v1.40.0
	github.com/aws/aws-sdk-go-v2/config v1.32.1
	github.com/aws/aws-sdk-go-v2/credentials v1.19.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.274.0
	github.com/aws/aws-sdk-go-v2/service/ssm v1.67.3
	github.com/aws/smithy-go v1.23.2
	github.com/danieljoos/wincred v1.2.3
	github.com/hashicorp/go-version v1.8.0
	github.com/spf13/cobra v1.10.1
)

require github.com/aws/aws-sdk-go-v2/service/signin v1.0.1 // indirect

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0
	golang.org/x/text v0.31.0
)
