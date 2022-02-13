import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as ecr from 'aws-cdk-lib/aws-ecr';

interface RunnerECRProps extends StackProps {
  application: string;
}

export class RunnerECR extends Stack {
  constructor(scope: Construct, id: string, props: RunnerECRProps) {
    super(scope, id, props);

    new ecr.Repository(this, 'ECR', {
      repositoryName: `${props.application}-ecr`,
      imageScanOnPush: true,
    });
  }
}
