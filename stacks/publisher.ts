import { join } from 'path';
import { Construct } from 'constructs';
import { Duration, RemovalPolicy, Stack, StackProps } from 'aws-cdk-lib';
import { Queue } from 'aws-cdk-lib/aws-sqs';
import {
  AttributeType,
  BillingMode,
  ProjectionType,
  StreamViewType,
  Table,
} from 'aws-cdk-lib/aws-dynamodb';
import { Code, Function, Runtime } from 'aws-cdk-lib/aws-lambda';
import { LambdaRestApi, MethodLoggingLevel } from 'aws-cdk-lib/aws-apigateway';
import { Rule, Schedule } from 'aws-cdk-lib/aws-events';
import { LambdaFunction } from 'aws-cdk-lib/aws-events-targets';

interface PublisherProps extends StackProps {
  application: string;
  githubAppID: string;
  githubAppPrivateKey: string;
  githubWebhookSecret: string;
  githubToken: string;
  ec2ConcurrencyLimit: number;
  eksConcurrencyLimit: number;
}

export class Publisher extends Stack {
  readonly jobsTableHostIndex: string = 'HostIndex';

  readonly launchQueueVisibilityTimeout: Duration = Duration.minutes(1);

  readonly terminationQueueVisibilityTimeout: Duration = Duration.seconds(30);

  readonly publisherRate: Duration = Duration.minutes(1);

  launchQueue: Queue;

  terminationQueue: Queue;

  constructor(scope: Construct, id: string, props: PublisherProps) {
    super(scope, id, props);

    const eventsTable = this.createJobsTable(
      props.application,
      this.jobsTableHostIndex
    );
    const producer = this.createProducer(
      props.application,
      props.githubAppID,
      props.githubAppPrivateKey,
      props.githubWebhookSecret,
      eventsTable.tableName,
      'debug'
    );

    new LambdaRestApi(this, 'PublisherAPIGateway', {
      handler: producer,
      deployOptions: {
        metricsEnabled: true,
        loggingLevel: MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
      },
    });

    this.launchQueue = new Queue(this, 'LaunchSQS', {
      queueName: `${props.application}-launch-sqs`,
      visibilityTimeout: this.launchQueueVisibilityTimeout,
    });

    this.terminationQueue = new Queue(this, 'TerminationSQS', {
      queueName: `${props.application}-termination-sqs`,
      visibilityTimeout: this.terminationQueueVisibilityTimeout,
    });

    const publisher = this.createPublisher(
      props.application,
      props.ec2ConcurrencyLimit,
      props.eksConcurrencyLimit,
      eventsTable.tableName,
      this.jobsTableHostIndex,
      this.launchQueue.queueUrl,
      this.terminationQueue.queueUrl
    );

    eventsTable.grantWriteData(producer);
    eventsTable.grantReadWriteData(publisher);

    new Rule(this, 'ScheduleRule', {
      schedule: Schedule.rate(this.publisherRate),
      targets: [new LambdaFunction(publisher)],
    });

    this.launchQueue.grantSendMessages(publisher);
    this.terminationQueue.grantSendMessages(publisher);
  }

  createProducer(
    application: string,
    appID: string,
    privateKey: string,
    webhookSecret: string,
    table: string,
    logLevel: string
    // eslint-disable-next-line @typescript-eslint/ban-types
  ): Function {
    return new Function(this, 'ProducerLambda', {
      functionName: `${application}-producer`,
      handler: 'lambda.handler',
      runtime: Runtime.NODEJS_14_X,
      code: Code.fromAsset(join(__dirname, '..', 'producer', 'dist')),
      memorySize: 512,
      timeout: Duration.seconds(30),
      environment: {
        APP_ID: appID,
        PRIVATE_KEY: privateKey,
        WEBHOOK_SECRET: webhookSecret,
        JOBS_TABLE: table,
        LOG_LEVEL: logLevel,
      },
    });
  }

  createPublisher(
    application: string,
    ec2Limits: number,
    eksLimits: number,
    table: string,
    index: string,
    launchQueue: string,
    terminationQueue: string
    // eslint-disable-next-line @typescript-eslint/ban-types
  ): Function {
    return new Function(this, 'PublisherLambda', {
      functionName: `${application}-publisher`,
      handler: 'publisher',
      runtime: Runtime.GO_1_X,
      code: Code.fromAsset(join(__dirname, '..', 'publisher', '_dist')),
      memorySize: 512,
      timeout: Duration.seconds(30),
      environment: {
        EC2_CURRENCY_LIMIT: ec2Limits.toString(),
        EKS_CURRENCY_LIMIT: eksLimits.toString(),
        JOBS_TABLE: table,
        JOBS_TABLE_HOST_INDEX: index,
        LAUNCH_QUEUE_URL: launchQueue,
        TERMINATION_QUEUE_URL: terminationQueue,
      },
    });
  }

  createJobsTable(application: string, index: string): Table {
    const table = new Table(this, 'JobsTable', {
      tableName: `${application}-jobs`,
      partitionKey: { name: 'ID', type: AttributeType.NUMBER },
      stream: StreamViewType.KEYS_ONLY,
      billingMode: BillingMode.PAY_PER_REQUEST,
      removalPolicy: RemovalPolicy.DESTROY,
    });

    table.addGlobalSecondaryIndex({
      indexName: index,
      partitionKey: { name: 'Host', type: AttributeType.STRING },
      sortKey: { name: 'CreatedAt', type: AttributeType.NUMBER },
      projectionType: ProjectionType.INCLUDE,
      nonKeyAttributes: ['Status', 'Content'],
    });

    return table;
  }
}
