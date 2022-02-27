import {
  DynamoDBClient,
  PutItemCommand,
  UpdateItemCommand,
} from '@aws-sdk/client-dynamodb';
import { promisify } from 'util';
import { gzip } from 'zlib';

export enum Host {
  EC2 = 'ec2',
  EKS = 'eks',
}

export enum Status {
  Queued = 'queued',
  Completed = 'completed',
}

export interface Job {
  id: number;
  owner: string;
  repository: string;
  labels: string[];
}

const compressJob = async (job: Job): Promise<Buffer> =>
  (await promisify(gzip)(Buffer.from(JSON.stringify(job)))) as Buffer;

const getFirstOS = (
  labels: string[],
  supportedOS: string[]
): string | undefined => labels.find((l: string) => supportedOS.includes(l));

export class Storage {
  private readonly client: DynamoDBClient;

  private readonly tableName: string;

  private readonly supportedOS: string[];

  constructor(
    client: DynamoDBClient,
    tableName: string,
    supportedOS: string[]
  ) {
    this.client = client;
    this.tableName = tableName;
    this.supportedOS = supportedOS;
  }

  async store(job: Job): Promise<void> {
    const os = getFirstOS(job.labels, this.supportedOS);
    if (!os) {
      throw new Error(`no supported OS found in labels (${job.labels})`);
    }

    try {
      await this.client.send(
        new PutItemCommand({
          TableName: this.tableName,
          ConditionExpression: 'attribute_not_exists(ID)',
          Item: {
            ID: { N: job.id.toString() },
            Host: {
              S: job.labels.includes(Host.EKS) ? Host.EKS : Host.EC2,
            },
            OS: {
              S: os,
            },
            Content: { B: await compressJob(job) },
            Status: { S: Status.Queued },
            CreatedAt: { N: Date.now().toString() },
          },
        })
      );
    } catch (e) {
      if (!e.name || e.name !== 'ConditionalCheckFailedException') {
        throw e;
      }
    }
  }

  async setJobCompleted(id: number): Promise<void> {
    await this.client.send(
      new UpdateItemCommand({
        TableName: this.tableName,
        Key: { ID: { N: id.toString() } },
        UpdateExpression: 'SET #s = :s',
        ConditionExpression: 'attribute_exists(ID)',
        ExpressionAttributeNames: {
          '#s': 'Status',
        },
        ExpressionAttributeValues: {
          ':s': { S: Status.Completed },
        },
      })
    );
  }
}
