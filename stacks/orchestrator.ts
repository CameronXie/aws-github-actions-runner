import { join } from 'path';
import { Duration, Stack, StackProps } from 'aws-cdk-lib';
import { Function, Runtime, Code } from 'aws-cdk-lib/aws-lambda';
import { Effect, Policy, PolicyStatement } from 'aws-cdk-lib/aws-iam';
import { SqsEventSource } from 'aws-cdk-lib/aws-lambda-event-sources';
import { Queue } from 'aws-cdk-lib/aws-sqs';
import { Construct } from 'constructs';

interface RunnerEKS {
  cluster: string;
  runnerNamespace: string;
  githubTokenSecret: string;
  githubTokenSecretKey: string;
}

interface Container {
  image: string;
  cpu: string;
  memory: string;
}

interface OrchestratorProps extends StackProps {
  application: string;
  githubToken: string;
  launchQueue: Queue;
  terminationQueue: Queue;
  ubuntuLaunchTemplateID: string;
  cluster: RunnerEKS;
  ubuntuRunnerContainer: Container;
  dindContainer: Container;
  subnetID: string;
  runnerVersion: string;
}

export class Orchestrator extends Stack {
  private readonly workflowJobID = 'GITHUB_WORKFLOW_JOB_ID';

  private readonly sqsBatchSize = 1;

  private readonly launchMemory: number = 128;

  private readonly launchQueueVisibilityTimeout: Duration = Duration.minutes(1);

  private readonly terminationQueueVisibilityTimeout: Duration =
    Duration.seconds(30);

  constructor(scope: Construct, id: string, props: OrchestratorProps) {
    super(scope, id, props);

    const launcher = this.createLauncher(
      props.application,
      props.ubuntuLaunchTemplateID,
      props.cluster,
      props.ubuntuRunnerContainer,
      props.dindContainer,
      props.subnetID,
      props.githubToken,
      props.runnerVersion,
      this.launchMemory,
      this.launchQueueVisibilityTimeout
    );

    launcher.addEventSource(
      new SqsEventSource(props.launchQueue, { batchSize: this.sqsBatchSize })
    );

    const terminator = this.createTerminator(
      props.application,
      props.cluster,
      this.launchMemory,
      this.terminationQueueVisibilityTimeout
    );
    terminator.addEventSource(
      new SqsEventSource(props.terminationQueue, {
        batchSize: this.sqsBatchSize,
      })
    );
  }

  createLauncher(
    application: string,
    ubuntuLaunchTemplateID: string,
    cluster: RunnerEKS,
    ubuntuRunnerContainer: Container,
    dindContainer: Container,
    subnetID: string,
    githubToken: string,
    runnerVersion: string,
    memorySize: number,
    timeout: Duration
    // eslint-disable-next-line @typescript-eslint/ban-types
  ): Function {
    const launcher = new Function(this, 'LauncherLambda', {
      functionName: `${application}-launcher`,
      handler: 'launcher',
      runtime: Runtime.GO_1_X,
      memorySize,
      timeout,
      code: Code.fromAsset(
        join(__dirname, '..', 'orchestrator', '_dist', 'launcher')
      ),
      environment: {
        // EKS
        EKS_CLUSTER: cluster.cluster,
        EKS_NAMESPACE: cluster.runnerNamespace,
        GITHUB_TOKEN_SECRET: cluster.githubTokenSecret,
        GITHUB_TOKEN_SECRET_TOKEN: cluster.githubTokenSecretKey,
        UBUNTU_RUNNER_CONTAINER_IMAGE: ubuntuRunnerContainer.image,
        UBUNTU_RUNNER_CONTAINER_CPU: ubuntuRunnerContainer.cpu,
        UBUNTU_RUNNER_CONTAINER_MEMORY: ubuntuRunnerContainer.memory,
        DIND_CONTAINER_IMAGE: dindContainer.image,
        DIND_CONTAINER_CPU: dindContainer.cpu,
        DIND_CONTAINER_MEMORY: dindContainer.memory,
        // EC2
        UBUNTU_LAUNCH_TEMPLATE_ID: ubuntuLaunchTemplateID,
        SUBNET_ID: subnetID,
        GITHUB_TOKEN: githubToken,
        GITHUB_RUNNER_VERSION: runnerVersion,
      },
    });

    launcher.role?.attachInlinePolicy(
      new Policy(this, 'LaunchInstance', {
        statements: [
          new PolicyStatement({
            actions: [
              'ec2:DescribeInstances',
              'ec2:DescribeImages',
              'ec2:DescribeInstanceTypes',
              'ec2:DescribeKeyPairs',
              'ec2:DescribeVpcs',
              'ec2:DescribeSubnets',
              'ec2:DescribeSecurityGroups',
              'eks:DescribeCluster',
              'ec2:CreateTags',
              'ec2:RunInstances',
            ],
            effect: Effect.ALLOW,
            resources: ['*'],
          }),
          new PolicyStatement({
            actions: ['ec2:RunInstances'],
            effect: Effect.DENY,
            resources: ['arn:aws:ec2:*:*:instance/*'],
            conditions: {
              'ForAllValues:StringNotEquals': {
                'aws:TagKeys': [this.workflowJobID],
              },
            },
          }),
          new PolicyStatement({
            actions: ['iam:PassRole'],
            effect: Effect.ALLOW,
            resources: ['*'],
            conditions: {
              StringEquals: { 'iam:PassedToService': ['ec2.amazonaws.com'] },
            },
          }),
        ],
      })
    );

    return launcher;
  }

  createTerminator(
    application: string,
    cluster: RunnerEKS,
    memorySize: number,
    timeout: Duration
    // eslint-disable-next-line @typescript-eslint/ban-types
  ): Function {
    const terminator = new Function(this, 'TerminatorLambda', {
      functionName: `${application}-terminator`,
      handler: 'terminator',
      runtime: Runtime.GO_1_X,
      code: Code.fromAsset(
        join(__dirname, '..', 'orchestrator', '_dist', 'terminator')
      ),
      memorySize,
      timeout,
      environment: {
        // EKS
        EKS_CLUSTER: cluster.cluster,
        EKS_NAMESPACE: cluster.runnerNamespace,
      },
    });

    terminator.role?.attachInlinePolicy(
      new Policy(this, 'TerminateInstance', {
        statements: [
          new PolicyStatement({
            actions: ['ec2:DescribeInstances', 'eks:DescribeCluster'],
            effect: Effect.ALLOW,
            resources: ['*'],
          }),
          new PolicyStatement({
            actions: ['ec2:TerminateInstances'],
            effect: Effect.ALLOW,
            resources: ['*'],
            conditions: {
              StringLike: { [`ec2:ResourceTag/${this.workflowJobID}`]: '*' },
            },
          }),
        ],
      })
    );

    return terminator;
  }
}
