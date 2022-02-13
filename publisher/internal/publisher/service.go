package publisher

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/queue"
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
	Queued    []storage.Job
	Completed []storage.Job
}

type Publisher interface {
	Publish(ctx context.Context) error
}

type publisher struct {
	storage             storage.Storage
	queue               queue.Queue
	hostOptions         []HostOption
	launchQueueURL      string
	terminationQueueURL string
	logger              *zap.Logger
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
		zap.Ints("queued", getJobIDs(jobs.Queued)),
		zap.Ints("completed", getJobIDs(jobs.Completed)),
	)

	msg := append(
		toQueueMessages(p.terminationQueueURL, jobs.Completed),
		toQueueMessages(p.launchQueueURL, jobs.Queued)...,
	)

	if len(msg) == 0 {
		return nil
	}

	qErr := p.queue.Send(ctx, msg)

	if qErr != nil {
		return qErr
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

	queued, err := p.storage.GetJobs(ctx, &storage.GetJobsInput{
		Host:     opt.Host,
		Statuses: []string{queuedStatus},
		Limit:    opt.Limit + int32(len(completed)),
	})

	if err != nil {
		return nil, err
	}

	return &Jobs{
		Queued:    queued,
		Completed: completed,
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

	d := make([]int, 0)
	for _, i := range jobs.Completed {
		d = append(d, i.ID)
	}

	return p.storage.UpdateJobs(ctx, &storage.UpdateJobsInput{
		Update: u,
		Delete: d,
	})
}

func toQueueMessages(url string, jobs []storage.Job) []queue.Message {
	res := make([]queue.Message, 0)
	for _, i := range jobs {
		b, _ := json.Marshal(i.Content)
		res = append(res, queue.Message{
			URL:             url,
			Body:            string(b),
			DeduplicationID: strconv.Itoa(i.ID),
		})
	}

	return res
}

func getJobIDs(jobs []storage.Job) []int {
	ids := make([]int, 0)
	for _, i := range jobs {
		ids = append(ids, i.ID)
	}

	return ids
}

func New(
	launchQueueURL string,
	terminationQueueURL string,
	hostOptions []HostOption,
	s storage.Storage,
	q queue.Queue,
	logger *zap.Logger,
) Publisher {
	return &publisher{
		launchQueueURL:      launchQueueURL,
		terminationQueueURL: terminationQueueURL,
		hostOptions:         hostOptions,
		storage:             s,
		queue:               q,
		logger:              logger,
	}
}
