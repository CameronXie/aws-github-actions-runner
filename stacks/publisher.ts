import { join } from 'path';
import { Construct } from 'constructs';
import { Duration, RemovalPolicy, Stack, StackProps } from 'aws-cdk-lib';
import {
  AttributeType,
  BillingMode,
  ProjectionType,
  StreamViewType,
  Table,
} from 'aws-cdk-lib/aws-dynamodb';
import {
  Code,
  Function,
  Runtime,
  StartingPosition,
} from 'aws-cdk-lib/aws-lambda';
import { LambdaRestApi, MethodLoggingLevel } from 'aws-cdk-lib/aws-apigateway';
import { DynamoEventSource } from 'aws-cdk-lib/aws-lambda-event-sources';
import { Topic } from 'aws-cdk-lib/aws-sns';
import { LambdaSubscription } from 'aws-cdk-lib/aws-sns-subscriptions';

interface PublisherProps extends StackProps {
  application: string;
  githubAppID: string;
  githubAppPrivateKey: string;
  githubWebhookSecret: string;
  githubToken: string;
  ec2ConcurrencyLimit: number;
  eksConcurrencyLimit: number;
  logLevel: string;
}

export class Publisher extends Stack {
  private readonly jobsTableHostIndex: string = 'HostIndex';

  private readonly dynamodbBatchSize: number = 10;

  private readonly dynamodbBatchingWindow: Duration = Duration.seconds(10);

  private readonly lambdaMemorySize: number = 512;

  private readonly lambdaDuration: Duration = Duration.seconds(30);

  jobsTopic: Topic;

  constructor(scope: Construct, id: string, props: PublisherProps) {
    super(scope, id, props);

    const publisherTopic = new Topic(this, 'PublisherTopic', {
      displayName: `${props.application}-publisher`,
    });

    this.jobsTopic = new Topic(this, 'JobsTopic', {
      displayName: `${props.application}-jobs`,
    });

    const jobsTable = this.createJobsTable(
      props.application,
      this.jobsTableHostIndex
    );

    const producer = this.createProducer(
      props.application,
      props.githubAppID,
      props.githubAppPrivateKey,
      props.githubWebhookSecret,
      jobsTable.tableName,
      props.logLevel
    );

    const messenger = this.createMessenger(
      props.application,
      publisherTopic.topicArn
    );

    const publisher = this.createPublisher(
      props.application,
      props.ec2ConcurrencyLimit,
      props.eksConcurrencyLimit,
      jobsTable.tableName,
      this.jobsTableHostIndex,
      publisherTopic.topicArn,
      this.jobsTopic.topicArn
    );

    new LambdaRestApi(this, 'PublisherAPIGateway', {
      handler: producer,
      deployOptions: {
        metricsEnabled: true,
        loggingLevel: MethodLoggingLevel.INFO,
        dataTraceEnabled: true,
      },
    });
    jobsTable.grantWriteData(producer);

    messenger.addEventSource(
      new DynamoEventSource(jobsTable, {
        startingPosition: StartingPosition.TRIM_HORIZON,
        batchSize: this.dynamodbBatchSize,
        maxBatchingWindow: this.dynamodbBatchingWindow,
        bisectBatchOnError: true,
        retryAttempts: 3,
      })
    );
    publisherTopic.grantPublish(messenger);

    publisherTopic.addSubscription(new LambdaSubscription(publisher));
    this.jobsTopic.grantPublish(publisher);
    jobsTable.grantReadWriteData(publisher);
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
      nonKeyAttributes: ['OS', 'Content', 'Status'],
    });

    return table;
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
      memorySize: this.lambdaMemorySize,
      timeout: this.lambdaDuration,
      environment: {
        APP_ID: appID,
        PRIVATE_KEY: privateKey,
        WEBHOOK_SECRET: webhookSecret,
        JOBS_TABLE: table,
        LOG_LEVEL: logLevel,
      },
    });
  }

  // eslint-disable-next-line @typescript-eslint/ban-types
  createMessenger(application: string, publisherTopic: string): Function {
    return new Function(this, 'MessengerLambda', {
      functionName: `${application}-messenger`,
      handler: 'messenger',
      runtime: Runtime.GO_1_X,
      code: Code.fromAsset(join(__dirname, '..', 'messenger', '_dist')),
      memorySize: this.lambdaMemorySize,
      timeout: this.lambdaDuration,
      environment: {
        PUBLISHER_TOPIC: publisherTopic,
      },
    });
  }

  createPublisher(
    application: string,
    ec2Limits: number,
    eksLimits: number,
    table: string,
    index: string,
    publisherTopic: string,
    jobsTopic: string
    // eslint-disable-next-line @typescript-eslint/ban-types
  ): Function {
    return new Function(this, 'PublisherLambda', {
      functionName: `${application}-publisher`,
      handler: 'publisher',
      runtime: Runtime.GO_1_X,
      code: Code.fromAsset(join(__dirname, '..', 'publisher', '_dist')),
      memorySize: this.lambdaMemorySize,
      reservedConcurrentExecutions: 1,
      timeout: this.lambdaDuration,
      environment: {
        EC2_CURRENCY_LIMIT: ec2Limits.toString(),
        EKS_CURRENCY_LIMIT: eksLimits.toString(),
        JOBS_TABLE: table,
        JOBS_TABLE_HOST_INDEX: index,
        PUBLISHER_TOPIC: publisherTopic,
        JOBS_TOPIC: jobsTopic,
      },
    });
  }
}
