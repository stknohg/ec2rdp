# ec2rdp

Remote Desktop Utility for Amazon EC2.  

This tool assists you to easily connet to your EC2 instances with Remote Desktop Client.

## Prerequisites

* Windows,  (Experimental) macOS
* [Session Manager plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)
* [Parallels Client](https://www.parallels.com/products/ras/capabilities/rdp-client/) 19+ is needed on macOS

### Required IAM actions

* `ec2:DescribeInstances`
* `ec2:GetPasswordData`
* `ssm:DescribeInstanceInformation`
* `ssm:StartSession`
* `ssm:TerminateSession`

## How to install

Download `ec2rdp` binary and setup [AWS CLI credential file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).

## Commands

### ec2rdp public

Connect to public EC2 instance with Remote Desktop Client.

```powershell
ec2rdp public -i 'EC2 instance ID' -p 'Path to private key file (.pem)'
```
#### example

```powershell
# Connect to EC2
PS C:\> $env:AWS_PROFILE='your_profile'
PS C:\> ec2rdp public -i i-01234567890abcdef -p C:\project\example.pem
```

### ec2rdp ssm

Connect to EC2 instance with Remote Desktop Client via SSM port forwarding.

```powershell
ec2rdp ssm -i 'EC2 instance ID' -p 'Path to private key file (.pem)'
```

#### example

```powershell
# Connect to EC2
PS C:\> $env:AWS_PROFILE='your_profile'
PS C:\> ec2rdp ssm -i i-01234567890abcdef -p C:\project\example.pem
```

### Customization

You can use `--profile`, `--region` parameters.

```powershell
PS C:\> ec2rdp public -i i-01234567890abcdef -p C:\project\example.pem --profile your_profile --region ap-northeast-1
```

You can override RDP connection settings by `--port`, `--user`, `--password` parameters.

```powershell
PS C:\> ec2rdp ssm -i i-01234567890abcdef --port 3390 --user MyAdmin --password 'P@ssword123456'

# It is recommended to use some kind of vault tool for better security
PS C:\> ec2rdp ssm -i i-01234567890abcdef --port 3390 --user MyAdmin --password $(Get-Secret -Name MyPassword -AsPlainText)
```

## License

* [MIT](./LICENSE)
