import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ec2 from 'aws-cdk-lib/aws-ec2';

interface RunnerTemplateProps extends StackProps {
  application: string;
  vpc: ec2.IVpc;
  ubuntuInstanceType: ec2.InstanceType;
}

export class RunnerTemplate extends Stack {
  private readonly ubuntuAmiName =
    'ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server*';

  private readonly ubuntuAmiOwner = '099720109477';

  ubuntuLaunchTemplate: ec2.LaunchTemplate;

  constructor(scope: Construct, id: string, props: RunnerTemplateProps) {
    super(scope, id, props);

    const role = new iam.Role(this, 'InstanceRole', {
      roleName: `${props.application}-instance-role`,
      assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
      description: `${props.application} Runner Instance Role`,
    });

    role.addManagedPolicy(
      iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMManagedInstanceCore')
    );

    const sg = new ec2.SecurityGroup(this, 'SecurityGroup', {
      vpc: props.vpc,
      allowAllOutbound: true,
    });

    this.ubuntuLaunchTemplate = this.createLaunchTemplate(
      'Ubuntu',
      props.application,
      this.ubuntuAmiName,
      [this.ubuntuAmiOwner],
      props.ubuntuInstanceType,
      role,
      sg
    );
  }

  createLaunchTemplate(
    templateName: string,
    application: string,
    amiName: string,
    amiOwners: string[],
    instanceType: ec2.InstanceType,
    role: iam.Role,
    sg: ec2.SecurityGroup
  ): ec2.LaunchTemplate {
    return new ec2.LaunchTemplate(this, `${templateName}Template`, {
      launchTemplateName: `${application}-template`,
      machineImage: ec2.MachineImage.lookup({
        name: amiName,
        owners: amiOwners,
      }),
      instanceType: instanceType,
      ebsOptimized: true,
      securityGroup: sg,
      role,
    });
  }
}
