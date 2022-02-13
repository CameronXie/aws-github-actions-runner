import { CfnOutput, Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as ec2 from 'aws-cdk-lib/aws-ec2';

interface VpcProps extends StackProps {
  CIDR: string;
}

export class Vpc extends Stack {
  vpc: ec2.IVpc;

  constructor(scope: Construct, id: string, props: VpcProps) {
    super(scope, id, props);

    this.vpc = new ec2.Vpc(this, 'VPC', {
      cidr: props.CIDR,
      natGateways: 2,
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: 'Public',
          subnetType: ec2.SubnetType.PUBLIC,
        },
        {
          cidrMask: 24,
          name: 'Private',
          subnetType: ec2.SubnetType.PRIVATE_WITH_NAT,
        },
      ],
    });

    new CfnOutput(this, 'vpc-id', {
      description: 'VPC Id',
      exportName: `${this.stackName}-vpc-id`,
      value: this.vpc.vpcId,
    });
  }
}
