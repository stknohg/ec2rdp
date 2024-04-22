# ec2rdp

Remote Desktop Utility for Amazon EC2.  

This tool assists you to easily connet to your EC2 instances with Remote Desktop Client.

## Prerequisites

* Windows, (Experimental) macOS
* [Parallels Client](https://www.parallels.com/products/ras/capabilities/rdp-client/) 19+ is needed on macOS
* (Optional) [AWS CLI](https://aws.amazon.com/cli/) 2.12.0+
    * Required when using `ec2rdp eice` command
* (Optional) [Session Manager plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)
    * Required when using `ec2rdp ssm` command

### Required IAM actions

* `ec2:DescribeInstances`
* `ec2:GetPasswordData`
* `ec2:DescribeInstanceConnectEndpoints`
* `ssm:DescribeInstanceInformation`
* `ssm:StartSession`
* `ssm:TerminateSession`
* `ec2-instance-connect:OpenTunnel`

## How to install

Download `ec2rdp` binary and setup [AWS CLI credential file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).  

### Windows

You can use [Scoop](https://scoop.sh/) on Windows.

```PowerShell
# Windows only
scoop bucket add stknohg https://github.com/stknohg/scoop-bucket
scoop install ec2rdp
```

### macOS

You can use Homebrew Taps on macOS.

```bash
# macOS only
brew install stknohg/tap/ec2rdp
```

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

### ec2rdp eice

Connect to EC2 instance with Remote Desktop Client via EC2 Instance Connect Endpoint.

```powershell
ec2rdp eice -i 'EC2 instance ID' -p 'Path to private key file (.pem)'
```

You can also use `--endpointid`(`-e`) flag to specify EC2 Instance Connect Endpoint ID.  

```powershell
ec2rdp eice -e 'Endpoint ID' -i 'EC2 instance ID' -p 'Path to private key file (.pem)'
```

#### example

```powershell
# Connect to EC2 via endpoint in the same VPC
PS C:\> $env:AWS_PROFILE='your_profile'
PS C:\> ec2rdp eice -i i-01234567890abcdef -p C:\project\example.pem

# Connect to EC2 via spcecified endpoint
PS C:\> ec2rdp eice -e eice-xxxxxxxxxx -i i-01234567890abcdef -p C:\project\example.pem
```

### Customization

You can use `--profile`, `--region` parameters.

```powershell
PS C:\> ec2rdp public -i i-01234567890abcdef -p C:\project\example.pem --profile your_profile --region ap-northeast-1
```

You can override RDP connection settings by `--port`, `--user`, `--password` parameters.

```powershell
PS C:\> ec2rdp ssm -i i-01234567890abcdef --port 3390 --user MyAdmin --password
```

## License

* [MIT](./LICENSE)
