# AWS GitHub Actions Runner

This project demonstrates an approach to orchestrate ephemeral GitHub Actions Runner with different hosting for
different workflow job. For example, using `container hosting runner` for high volume code scanning or linting workflow,
and using `virtual machine hosting runner` for workflow which difficult to run on container, like iOS Apps build.

## Architecture

### Components

* `Producter` An NodeJS Lambda function which stores `workflow_job` events in `Jobs Table`.
* `Jobs Table` An DynamoDB Table which tracks `workflow_job` status.
* `Publisher` An Golang Lambda function which queries `Jobs Table` and sends runner orchestration message to
  `Jobs Topic` SNS. `Publisher` responsible for the `flow control`, we don't want overheating AWS resource.
* `Orchestrator (SQS + Lambda)` An SQS subscribes `Jobs Topic` on one particular runner host and OS combination (
  example: EC2 and ubuntu subscription filter policy), and triggers a lambda to perform orchestration, spin up or
  tear down runners.

### Diagram

![AWS GitHub Actions Runner](./images/aws-github-actions-runner.png "AWS GitHub Actions Runner")

1. GitHub sends `workflow_job` events to `Producer`.
2. `Producer` inserts job events into `Jobs Table` with status ('queued' or 'completed').
3. `Jobs Table` DynamoDB stream invokes `Messenger` lambda to publish a notification on `Publisher Topic`.
4. `Publisher` queries `Jobs Table` to get a limited number of jobs (FIFO).
5. `Publisher` publishes jobs to `Jobs Topic`.
6. `Publisher` also publishes a notification on `Publisher Topic` until there is no job remains in `Jobs Table`.
7. `Orchestrator` SQS subscribes `Jobs Topic` on one particular runner host and OS combination (e.g. eks ubuntu or ec2
   windows).
8. `Orchestrator` SQS triggers a lambda to perform orchestration operation, spin up or tear down.
9. Self-hosted runner will register in GitHub, and start polling queued `workflow_job`.

## Deploy

### Prerequisites

* An AWS IAM user account which has enough permission to deploy:
    * VPC (Subnets, Route Tables, NAT Gateway...etc.)
    * API Gateway
    * DynamoDB
    * Lambda
    * SNS
    * SQS
    * EC2
    * EKS
    * ECR
    * CloudWatch Events
* Set up a GitHub Apps in your GitHub Account which has enough permission to send `workflow_job` events, and save
  the `app id`, `private key`, `app secret` and `github token` in `.env` file.

### Deploy with Docker

* Run `docker compose run --rm deployer make ci-deploy` to deploy the solution.
* Update the GitHub App URL with the API Gateway endpoint.

## Test

* Run `docker compose run --rm deployer make test` to test:
    * Producer
    * Publisher
    * Orchestrator

## TODO

* Support Lambda hosting runner.
* Support Distributed Tracing.


