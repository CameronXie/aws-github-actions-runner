import { join } from 'path';
import { Duration, Stack, StackProps } from 'aws-cdk-lib';
import { Function, Runtime, Code } from 'aws-cdk-lib/aws-lambda';
import { Effect, Policy, PolicyStatement } from 'aws-cdk-lib/aws-iam';
import { SqsEventSource } from 'aws-cdk-lib/aws-lambda-event-sources';
import { Queue } from 'aws-cdk-lib/aws-sqs';
import { Construct } from 'constructs';
import { SubscriptionFilter, Topic } from 'aws-cdk-lib/aws-sns';
import { SqsSubscription } from 'aws-cdk-lib/aws-sns-subscriptions';
import * as sns from 'aws-cdk-lib/aws-sns';

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

type LambdaEnv = { [key: string]: string };

type SNSFilterPolicy = { [attribute: string]: sns.SubscriptionFilter };

interface OrchestratorProps extends StackProps {
  application: string;
  githubToken: string;
  jobsTopic: Topic;
  ubuntuLaunchTemplateID: string;
  cluster: RunnerEKS;
  ubuntuRunnerContainer: Container;
  dindContainer: Container;
  subnetID: string;
  runnerVersion: string;
}

enum Host {
  EC2 = 'ec2',
  EKS = 'eks',
}

enum OS {
  Ubuntu = 'ubuntu',
}

enum Status {
  Queued = 'queued',
  Completed = 'completed',
}

enum OrchestratorRole {
  Launcher = 'launcher',
  Terminator = 'terminator',
}

const capitalize = (word: string): string =>
  word.charAt(0).toUpperCase() + word.toLocaleLowerCase().slice(1);

export class Orchestrator extends Stack {
  private readonly workflowJobID = 'GITHUB_WORKFLOW_JOB_ID';

  private readonly lambdaMemory: number = 512;

  constructor(scope: Construct, id: string, props: OrchestratorProps) {
    super(scope, id, props);

    const ec2Orchestrator = this.createSQSLambdaSubscriber(
      props.application,
      Host.EC2
    );

    const ec2LauncherEnv = {
      SUBNET_ID: props.subnetID,
      GITHUB_TOKEN: props.githubToken,
      GITHUB_RUNNER_VERSION: props.runnerVersion,
    };

    const eksOrchestrator = this.createSQSLambdaSubscriber(
      props.application,
      Host.EKS
    );

    const eksLauncherEnv = {
      EKS_CLUSTER: props.cluster.cluster,
      EKS_NAMESPACE: props.cluster.runnerNamespace,
      GITHUB_TOKEN_SECRET: props.cluster.githubTokenSecret,
      GITHUB_TOKEN_SECRET_TOKEN: props.cluster.githubTokenSecretKey,
      DIND_CONTAINER_IMAGE: props.dindContainer.image,
      DIND_CONTAINER_CPU: props.dindContainer.cpu,
      DIND_CONTAINER_MEMORY: props.dindContainer.memory,
    };

    // EC2 Ubuntu Launcher
    props.jobsTopic.addSubscription(
      new SqsSubscription(
        ec2Orchestrator(
          OrchestratorRole.Launcher,
          this.getEC2LauncherPolicyStatements(),
          this.lambdaMemory,
          join(
            __dirname,
            '..',
            'orchestrator',
            '_dist',
            'launcher',
            'ec2',
            'ubuntu'
          ),
          Duration.minutes(1),
          {
            ...ec2LauncherEnv,
            LAUNCH_TEMPLATE_ID: props.ubuntuLaunchTemplateID,
          },
          OS.Ubuntu
        ),
        {
          filterPolicy: this.createSNSFilterPolicy(
            Host.EC2,
            Status.Queued,
            OS.Ubuntu
          ),
        }
      )
    );

    // EKS Ubuntu Launcher
    props.jobsTopic.addSubscription(
      new SqsSubscription(
        eksOrchestrator(
          OrchestratorRole.Launcher,
          this.getEKSOrchestratorPolicyStatements(),
          this.lambdaMemory,
          join(__dirname, '..', 'orchestrator', '_dist', 'launcher', 'eks'),
          Duration.minutes(1),
          {
            ...eksLauncherEnv,
            RUNNER_CONTAINER_IMAGE: props.ubuntuRunnerContainer.image,
            RUNNER_CONTAINER_CPU: props.ubuntuRunnerContainer.cpu,
            RUNNER_CONTAINER_MEMORY: props.ubuntuRunnerContainer.memory,
          },
          OS.Ubuntu
        ),
        {
          filterPolicy: this.createSNSFilterPolicy(
            Host.EKS,
            Status.Queued,
            OS.Ubuntu
          ),
        }
      )
    );

    // EC2 Terminator
    props.jobsTopic.addSubscription(
      new SqsSubscription(
        ec2Orchestrator(
          OrchestratorRole.Terminator,
          this.getEC2TerminatorPolicyStatements(),
          this.lambdaMemory,
          join(__dirname, '..', 'orchestrator', '_dist', 'terminator', 'ec2'),
          Duration.minutes(1)
        ),
        {
          filterPolicy: this.createSNSFilterPolicy(Host.EC2, Status.Completed),
        }
      )
    );

    // EKS Terminator
    props.jobsTopic.addSubscription(
      new SqsSubscription(
        eksOrchestrator(
          OrchestratorRole.Terminator,
          this.getEKSOrchestratorPolicyStatements(),
          this.lambdaMemory,
          join(__dirname, '..', 'orchestrator', '_dist', 'terminator', 'eks'),
          Duration.minutes(1),
          {
            EKS_CLUSTER: props.cluster.cluster,
            EKS_NAMESPACE: props.cluster.runnerNamespace,
          }
        ),
        {
          filterPolicy: this.createSNSFilterPolicy(Host.EKS, Status.Completed),
        }
      )
    );
  }

  createSNSFilterPolicy(host: Host, status: Status, os?: OS): SNSFilterPolicy {
    const policy: SNSFilterPolicy = {
      Host: SubscriptionFilter.stringFilter({
        allowlist: [host],
      }),
      Status: SubscriptionFilter.stringFilter({
        allowlist: [status],
      }),
    };

    if (os) {
      policy.OS = SubscriptionFilter.stringFilter({
        allowlist: [os],
      });
    }

    return policy;
  }

  createSQSLambdaSubscriber(
    application: string,
    host: Host
  ): (
    orchestratorRole: OrchestratorRole,
    lambdaPolicyStatements: PolicyStatement[],
    memorySize: number,
    codeFile: string,
    timeout: Duration,
    envs?: LambdaEnv,
    os?: OS
  ) => Queue {
    return (
      orchestratorRole: OrchestratorRole,
      lambdaPolicyStatements: PolicyStatement[],
      memorySize: number,
      codeFile: string,
      timeout: Duration,
      envs?: LambdaEnv,
      os?: OS
    ) => {
      const idPrefix = os
        ? host.toUpperCase() + capitalize(os) + capitalize(orchestratorRole)
        : host.toUpperCase() + capitalize(orchestratorRole);
      const name = os
        ? `${application}-${host}-${os}-${orchestratorRole}`.toLowerCase()
        : `${application}-${host}-${orchestratorRole}`.toLowerCase();

      const queue = new Queue(this, `${idPrefix}SQS`, {
        queueName: name,
        visibilityTimeout: timeout,
      });

      const lambda = new Function(this, `${idPrefix}Lambda`, {
        functionName: name,
        handler: orchestratorRole,
        runtime: Runtime.GO_1_X,
        memorySize,
        timeout,
        code: Code.fromAsset(codeFile),
        environment: envs,
      });

      lambda.role?.attachInlinePolicy(
        new Policy(this, `${idPrefix}Policy`, {
          statements: lambdaPolicyStatements,
        })
      );

      lambda.addEventSource(new SqsEventSource(queue, { batchSize: 1 }));
      return queue;
    };
  }

  getEC2LauncherPolicyStatements(): PolicyStatement[] {
    return [
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
    ];
  }

  getEC2TerminatorPolicyStatements(): PolicyStatement[] {
    return [
      new PolicyStatement({
        actions: ['ec2:DescribeInstances'],
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
    ];
  }

  getEKSOrchestratorPolicyStatements(): PolicyStatement[] {
    return [
      new PolicyStatement({
        actions: ['eks:DescribeCluster'],
        effect: Effect.ALLOW,
        resources: ['*'],
      }),
    ];
  }
}
