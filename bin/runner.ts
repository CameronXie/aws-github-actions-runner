#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Vpc } from '../stacks/vpc';
import { Orchestrator } from '../stacks/orchestrator';
import { RunnerTemplate } from '../stacks/runner-template';
import { RunnerECR } from '../stacks/runner-ecr';
import { Publisher } from '../stacks/publisher';

const getEnvStr = (key: string): string => {
  const v = process.env[key];
  if (v !== undefined) {
    return v;
  }

  throw new Error(`environment variable ${key} is not set`);
};

/*
 * EKS Runner Configuration
 *
 * configuration should align with create_eks.sh file.
 */
const cluster = 'actions-runner-cluster';
const runnerNamespace = 'actions-runner';
const githubTokenSecret = 'github-token';
const githubTokenSecretKey = 'GITHUB_PERSONAL_TOKEN';
const ubuntuRunnerContainer = {
  image: `${getEnvStr('CDK_DEFAULT_ACCOUNT')}.dkr.ecr.${getEnvStr(
    'CDK_DEFAULT_REGION'
  )}.amazonaws.com/actions-runner-ecr:latest`,
  cpu: '1',
  memory: '1Gi',
};
const dindContainer = {
  image: 'docker:dind',
  cpu: '500m',
  memory: '1Gi',
};

/*
 * EC2 Runner Configuration
 */
const ubuntuInstanceType = ec2.InstanceType.of(
  ec2.InstanceClass.T3,
  ec2.InstanceSize.MEDIUM
);

/*
 * Runner Configuration
 *
 * Modify EC2 and EKS Concurrency Limit.
 */
const application = 'actions-runner';
const ec2ConcurrencyLimit = 20;
const eksConcurrencyLimit = 20;
const env = {
  account: process.env.CDK_DEFAULT_ACCOUNT,
  region: process.env.CDK_DEFAULT_REGION,
};
const app = new cdk.App();
const vpc = new Vpc(app, `${application}-vpc`, {
  CIDR: '10.1.0.0/16',
  env,
});

new RunnerECR(app, `${application}-ecr`, {
  application,
  env,
});

const template = new RunnerTemplate(app, `${application}-template`, {
  application,
  vpc: vpc.vpc,
  ubuntuInstanceType,
  env,
});

template.addDependency(vpc);

const publisher = new Publisher(app, `${application}-publisher`, {
  application,
  githubAppID: getEnvStr('GITHUB_APP_ID'),
  githubAppPrivateKey: getEnvStr('GITHUB_APP_PRIVATE_KEY'),
  githubWebhookSecret: getEnvStr('GITHUB_APP_SECRET'),
  githubToken: getEnvStr('GITHUB_TOKEN'),
  ec2ConcurrencyLimit,
  eksConcurrencyLimit,
  env,
});

new Orchestrator(app, `${application}-orchestrator`, {
  application,
  githubToken: getEnvStr('GITHUB_TOKEN'),
  launchQueue: publisher.launchQueue,
  terminationQueue: publisher.terminationQueue,
  ubuntuLaunchTemplateID: template.ubuntuLaunchTemplate.launchTemplateId || '',
  cluster: {
    cluster,
    runnerNamespace,
    githubTokenSecret,
    githubTokenSecretKey,
  },
  ubuntuRunnerContainer,
  dindContainer,
  subnetID: vpc.vpc.selectSubnets({
    subnetType: ec2.SubnetType.PRIVATE_WITH_NAT,
  }).subnetIds[0],
  runnerVersion: getEnvStr('RUNNER_VERSION'),
  env,
});

app.synth();
