import {
  createLambdaFunction,
  createProbot,
} from '@probot/adapter-aws-lambda-serverless';
import bot from './bot';
import { Storage } from './storage';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';

const supportedOS = ['ubuntu'];

const getStorage = (): Storage => {
  console.log('init dynamodb client');
  const region = process.env.AWS_REGION;
  const table = process.env.JOBS_TABLE;

  if (!region || !table) {
    throw new Error('region or dynamodb table is incorrect');
  }

  return new Storage(new DynamoDBClient({ region }), table, supportedOS);
};

exports.handler = createLambdaFunction(bot(getStorage()), {
  probot: createProbot(),
});
