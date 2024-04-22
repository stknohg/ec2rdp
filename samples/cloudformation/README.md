# CloudFormation templates

## Sample IAM policies to connect EC2 instances

* For All regions : [iam-policy-all-regions.yaml](./iam-policy-all-regions.yaml)
* For a single region :  [iam-policy-single-region.yaml](./iam-policy-single-region.yaml)

These sample IAM policies grant the following privileges to the user.

* `ec2:DescribeInstances`
* `ec2:GetPasswordData`
* `ec2:DescribeInstanceConnectEndpoints`
* `ec2-instance-connect:OpenTunnel`
* `ssm:DescribeInstanceInformation`
* `ssm:StartSession`
* `ssm:TerminateSession`

As an example, [iam-policy-all-regions.yaml](./iam-policy-all-regions.yaml) will generate the following policy.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "EC2InstancesAndEndpoints",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeInstanceConnectEndpoints",
                "ec2:GetPasswordData"
            ],
            "Resource": "*"
        },
        {
            "Sid": "EICEOpenTunnel",
            "Effect": "Allow",
            "Action": [
                "ec2-instance-connect:OpenTunnel"
            ],
            "Resource": "arn:aws:ec2:*:<Account ID>:instance-connect-endpoint/*"
        },
        {
            "Sid": "SSMInstances",
            "Effect": "Allow",
            "Action": [
                "ssm:DescribeInstanceInformation"
            ],
            "Resource": "arn:aws:ssm:*:<Account ID>:*"
        },
        {
            "Sid": "SSMStartSession",
            "Effect": "Allow",
            "Action": [
                "ssm:StartSession"
            ],
            "Resource": [
                "arn:aws:ec2:*:<Account ID>:instance/*",
                "arn:aws:ssm:*::document/AWS-StartPortForwardingSession"
            ]
        },
        {
            "Sid": "SSMTerminateSession",
            "Effect": "Allow",
            "Action": [
                "ssm:TerminateSession",
                "ssm:ResumeSession"
            ],
            "Resource": [
                "arn:aws:ssm:*:<Account ID>:session/${aws:username}-*"
            ]
        }
    ]
}
```

Run the following command to apply these templates.

```bash
# For All regions
aws cloudformation create-stack --stack-name ec2rdp-connection-policy --template-body file://./iam-policy-all-regions.yaml --capabilities CAPABILITY_NAMED_IAM

# For a single region
aws cloudformation create-stack --stack-name ec2rdp-single-region-connection-policy --template-body file://./iam-policy-single-region.yaml --capabilities CAPABILITY_NAMED_IAM
```
