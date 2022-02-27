package publisher

import (
	"context"
	"encoding/json"

	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/messenger"
	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	queuedStatus     = "queued"
	inProgressStatus = "in_progress"
	completedStatus  = "completed"
)

type HostOption struct {
	Host  string
	Limit int32
}

type Jobs struct {
	Queued     []storage.Job
	InProgress []storage.Job
	Completed  []storage.Job
}

type Publisher interface {
	Publish(ctx context.Context) error
}

type publisher struct {
	hostOptions []HostOption
	messenger   messenger.Messenger
	storage     storage.Storage
	logger      *zap.Logger
}

func (p *publisher) Publish(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	for _, i := range p.hostOptions {
		opt := i
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
				return p.process(gCtx, opt)
			}
		})
	}

	return g.Wait()
}

func (p *publisher) process(ctx context.Context, opt HostOption) error {
	p.logger.Info("retrieving jobs",
		zap.String("host", opt.Host),
		zap.Int32("limit", opt.Limit),
	)

	jobs, jErr := p.getJobs(ctx, opt)

	if jErr != nil {
		return jErr
	}

	p.logger.Info("processing jobs",
		zap.Uint64s("queued", getJobIDs(jobs.Queued)),
		zap.Uint64s("in_progress", getJobIDs(jobs.InProgress)),
		zap.Uint64s("completed", getJobIDs(jobs.Completed)),
	)

	if len(jobs.Queued) != 0 || len(jobs.InProgress) != 0 || len(jobs.Completed) != 0 {
		defer func() {
			p.logger.Info(
				"notify publisher",
			)
			_ = p.messenger.NotifyPublisher(ctx)
		}()
	}

	if len(jobs.Queued) == 0 && len(jobs.Completed) == 0 {
		return nil
	}

	msg := append(toMessage(jobs.Queued), toMessage(jobs.Completed)...)
	nErr := p.messenger.PublishJobs(ctx, msg)
	if nErr != nil {
		return nErr
	}

	return p.updateJobs(ctx, *jobs)
}

func (p *publisher) getJobs(ctx context.Context, opt HostOption) (*Jobs, error) {
	completed, err := p.storage.GetJobs(ctx, &storage.GetJobsInput{
		Host:     opt.Host,
		Statuses: []string{completedStatus},
		Limit:    opt.Limit,
	})

	if err != nil {
		return nil, err
	}

	jobs, err := p.storage.GetJobs(ctx, &storage.GetJobsInput{
		Host:     opt.Host,
		Statuses: []string{queuedStatus, inProgressStatus},
		Limit:    opt.Limit + int32(len(completed)),
	})

	if err != nil {
		return nil, err
	}

	queued := make([]storage.Job, 0)
	inProgress := make([]storage.Job, 0)

	for i := range jobs {
		tmp := jobs[i]
		if tmp.Status == queuedStatus {
			queued = append(queued, tmp)
			continue
		}

		if tmp.Status == inProgressStatus {
			inProgress = append(inProgress, tmp)
		}
	}

	return &Jobs{
		Queued:     queued,
		InProgress: inProgress,
		Completed:  completed,
	}, nil
}

func (p *publisher) updateJobs(ctx context.Context, jobs Jobs) error {
	u := make([]storage.UpdateJob, 0)
	for _, i := range jobs.Queued {
		u = append(u, storage.UpdateJob{
			ID:     i.ID,
			Status: inProgressStatus,
		})
	}

	d := make([]uint64, 0)
	for _, i := range jobs.Completed {
		d = append(d, i.ID)
	}

	return p.storage.UpdateJobs(ctx, &storage.UpdateJobsInput{
		Update: u,
		Delete: d,
	})
}

func toMessage(jobs []storage.Job) []messenger.Message {
	res := make([]messenger.Message, 0)
	for _, i := range jobs {
		b, _ := json.Marshal(i.Content)
		res = append(res, messenger.Message{
			Host:   i.Host,
			OS:     i.OS,
			Status: i.Status,
			Body:   string(b),
		})
	}

	return res
}

func getJobIDs(jobs []storage.Job) []uint64 {
	ids := make([]uint64, 0)
	for _, i := range jobs {
		ids = append(ids, i.ID)
	}

	return ids
}

func New(
	s storage.Storage,
	m messenger.Messenger,
	opts []HostOption,
	logger *zap.Logger,
) Publisher {
	return &publisher{
		hostOptions: opts,
		storage:     s,
		messenger:   m,
		logger:      logger,
	}
}
