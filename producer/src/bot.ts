import { Context, Probot } from 'probot';
// eslint-disable-next-line  import/no-extraneous-dependencies
import { EmitterWebhookEvent } from '@octokit/webhooks/dist-types/types';
import { ApplicationFunction } from 'probot/lib/types';
import { Storage, Job } from './storage';

type RawData = { runnerName: string | null } & Job;

const getWorkflowJob = (
  context: EmitterWebhookEvent<'workflow_job'> & Context
): RawData => {
  const {
    payload: {
      repository,
      workflow_job: { id, labels, runner_name: runnerName },
    },
  } = context;

  return {
    id,
    owner: repository.owner.login,
    repository: repository.name,
    labels,
    runnerName,
  };
};

const getJobLog = (event: string, job: Job): string =>
  JSON.stringify({ event, job });

export default (storage: Storage): ApplicationFunction => {
  const RunnerNameSeparator = '-';

  return (bot: Probot) => {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    bot.on('workflow_job.queued', async (context) => {
      const raw = getWorkflowJob(context);
      context.log.info(getJobLog('workflow_job.queued', raw));

      const { runnerName, ...job } = raw;
      await storage.store(job);
    });

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    bot.on('workflow_job.completed', async (context) => {
      const raw = getWorkflowJob(context);
      context.log.info(getJobLog('workflow_job.completed', raw));
      if (!raw.runnerName) {
        await storage.setJobCompleted(raw.id);
        return;
      }

      const idStr = raw.runnerName
        .toString()
        .split(RunnerNameSeparator)
        .slice(-1)
        .pop();
      if (!idStr) {
        return;
      }

      const id = parseInt(idStr);
      if (!id) {
        return;
      }

      await storage.setJobCompleted(id);
    });
  };
};
