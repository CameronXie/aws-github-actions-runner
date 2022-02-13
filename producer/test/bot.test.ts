import nock = require('nock');
import producer from '../src/bot';
import { Probot } from 'probot';
import { Storage } from '../src/storage';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import pino from 'pino';
import * as Stream from 'stream';
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import queuedPayload from './fixtures/workflow_job.queued.json';
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import completedPayload from './fixtures/workflow_job.completed.json';
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import cancelledPayload from './fixtures/workflow_job.completed.cancelled.json';

const mockedStorage = {
  store: jest.fn(),
  setJobCompleted: jest.fn(),
};
jest.mock('../src/storage', () => ({
  ...jest.requireActual('../src/storage'),
  Storage: jest.fn(() => mockedStorage),
}));

let output: { msg: string }[] = [];
const streamLogsToOutput = new Stream.Writable({ objectMode: true });
streamLogsToOutput._write = (object, encoding, done) => {
  output.push(JSON.parse(object));
  done();
};

describe('Producer tests', () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let probot: any;

  beforeEach(() => {
    nock.disableNetConnect();
    probot = new Probot({
      githubToken: 'pat',
      log: pino(streamLogsToOutput),
    });
    output = [];
    producer(new Storage(new DynamoDBClient({}), ''))(probot, {});
  });

  it('should store the event when a workflow job is queued', async () => {
    mockedStorage.store.mockResolvedValue({});
    const expectedJob = {
      id: 2832853555,
      owner: 'octo-org',
      repository: 'example-workflow',
      labels: ['self-hosted', 'ubuntu', 'ec2'],
    };

    await probot.receive({
      name: 'workflow_job',
      payload: queuedPayload,
    });

    expect(mockedStorage.store).toBeCalledWith(expectedJob);
    expect(output[0].msg).toBe(
      JSON.stringify({
        event: 'workflow_job.queued',
        job: { ...expectedJob, runnerName: null },
      })
    );
  });

  it('should delete the event when a workflow job is completed', async () => {
    mockedStorage.setJobCompleted.mockResolvedValue({});
    const expectedJob = {
      id: 2832853555,
      owner: 'octo-org',
      repository: 'example-workflow',
      labels: ['self-hosted', 'ubuntu', 'ec2'],
      runnerName: 2832586764,
    };

    await probot.receive({
      name: 'workflow_job',
      payload: completedPayload,
    });

    expect(mockedStorage.setJobCompleted).toBeCalledWith(2832586764);
    expect(output[0].msg).toBe(
      JSON.stringify({ event: 'workflow_job.completed', job: expectedJob })
    );
  });

  it('should delete the event when a workflow job is cancelled', async () => {
    mockedStorage.setJobCompleted.mockResolvedValue({});
    const expectedJob = {
      id: 2832853555,
      owner: 'octo-org',
      repository: 'example-workflow',
      labels: ['self-hosted', 'ubuntu', 'ec2'],
      runnerName: null,
    };

    await probot.receive({
      name: 'workflow_job',
      payload: cancelledPayload,
    });

    expect(mockedStorage.setJobCompleted).toBeCalledWith(2832853555);
    expect(output[0].msg).toBe(
      JSON.stringify({ event: 'workflow_job.completed', job: expectedJob })
    );
  });

  afterEach(() => {
    nock.cleanAll();
    nock.enableNetConnect();
    mockedStorage.store.mockReset();
    mockedStorage.setJobCompleted.mockReset();
  });

  afterAll(() => {
    jest.clearAllMocks();
  });
});
