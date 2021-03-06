import { Status, Storage, Host, Job } from '../src/storage';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { promisify } from 'util';
import { gzip } from 'zlib';

const mockedClient = {
  send: jest.fn(),
};

jest.mock('@aws-sdk/client-dynamodb', () => ({
  ...jest.requireActual('@aws-sdk/client-dynamodb'),
  DynamoDBClient: jest.fn(() => mockedClient),
}));

const compressJob = async (job: Job): Promise<Buffer> =>
  (await promisify(gzip)(Buffer.from(JSON.stringify(job)))) as Buffer;

describe('Storage tests', () => {
  beforeEach(() => {
    jest.restoreAllMocks();
  });

  it('should store the job', async () => {
    mockedClient.send.mockResolvedValue({});
    const table = 'table';
    const supportedOS = ['ubuntu'];
    const job = {
      id: 123,
      labels: ['ubuntu'],
      owner: 'owner',
      repository: 'repo',
    };
    await new Storage(new DynamoDBClient({}), table, supportedOS).store(job);

    expect(mockedClient.send).toBeCalledWith(
      expect.objectContaining({
        input: expect.objectContaining({
          TableName: table,
          ConditionExpression: 'attribute_not_exists(ID)',
          Item: expect.objectContaining({
            ID: { N: '123' },
            Host: { S: Host.EC2 },
            OS: { S: 'ubuntu' },
            Status: { S: Status.Queued },
            Content: { B: await compressJob(job) },
            CreatedAt: { N: expect.anything() },
          }),
        }),
      })
    );
  });

  it('should throw the unsupported OS error', async () => {
    mockedClient.send.mockResolvedValue({});
    const table = 'table';
    const supportedOS = ['ubuntu'];
    const job = {
      id: 123,
      labels: ['windows'],
      owner: 'owner',
      repository: 'repo',
    };

    await expect(
      new Storage(new DynamoDBClient({}), table, supportedOS).store(job)
    ).rejects.toEqual(new Error('no supported OS found in labels (windows)'));
  });

  it('should set the job completed', async () => {
    mockedClient.send.mockResolvedValue({});
    const table = 'table';
    const supportedOS = ['ubuntu'];
    const id = 1234;
    await new Storage(
      new DynamoDBClient({}),
      table,
      supportedOS
    ).setJobCompleted(id);

    expect(mockedClient.send).toBeCalledWith(
      expect.objectContaining({
        input: expect.objectContaining({
          TableName: table,
          Key: { ID: { N: id.toString() } },
          UpdateExpression: 'SET #s = :s',
          ConditionExpression: 'attribute_exists(ID)',
          ExpressionAttributeNames: {
            '#s': 'Status',
          },
          ExpressionAttributeValues: {
            ':s': { S: Status.Completed },
          },
        }),
      })
    );
  });

  afterAll(() => {
    jest.clearAllMocks();
  });
});
