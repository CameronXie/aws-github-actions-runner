package publisher

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/messenger"
	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestPublisher_Publish(t *testing.T) {
	cases := map[string]struct {
		opt                []HostOption
		getJobErr          error
		updateJobErr       error
		publishJobsErr     error
		notifyPublisherErr error
		expectedLogs       []map[string]interface{}
		err                error
	}{
		"publish queued jobs": {
			opt: []HostOption{
				{Host: "ec2", Limit: 2},
				{Host: "eks", Limit: 2},
			},
			expectedLogs: []map[string]interface{}{
				{"msg": "retrieving jobs", "host": "ec2", "limit": int32(2)},
				{"msg": "processing jobs", "queued": []interface{}{uint64(1)}, "in_progress": []interface{}{uint64(2)}, "completed": []interface{}{}},
				{"msg": "notify publisher"},
				{"msg": "retrieving jobs", "host": "eks", "limit": int32(2)},
				{"msg": "processing jobs", "queued": []interface{}{uint64(4)}, "in_progress": []interface{}{}, "completed": []interface{}{uint64(5)}},
				{"msg": "notify publisher"},
			},
		},
		"not jobs found": {
			opt: []HostOption{
				{Host: "ec2", Limit: 0},
			},
			expectedLogs: []map[string]interface{}{
				{"msg": "retrieving jobs", "host": "ec2", "limit": int32(0)},
				{"msg": "processing jobs", "queued": []interface{}{}, "in_progress": []interface{}{}, "completed": []interface{}{}},
			},
		},
		"failed to get jobs": {
			opt: []HostOption{
				{Host: "ec2", Limit: 2},
				{Host: "eks", Limit: 2},
			},
			getJobErr: errors.New("failed to get jobs"),
			err:       errors.New("failed to get jobs"),
		},
		"failed to send jobs": {
			opt: []HostOption{
				{Host: "ec2", Limit: 2},
				{Host: "eks", Limit: 2},
			},
			publishJobsErr: errors.New("failed to send jobs"),
			err:            errors.New("failed to send jobs"),
		},
		"failed to update jobs": {
			opt: []HostOption{
				{Host: "ec2", Limit: 2},
				{Host: "eks", Limit: 2},
			},
			publishJobsErr: errors.New("failed to update jobs"),
			err:            errors.New("failed to update jobs"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			s := &mockedStorage{
				jobs:          getTestJobs(),
				getJobsErr:    tc.getJobErr,
				updateJobsErr: tc.updateJobErr,
			}
			m := &mockedMessenger{
				publishJobsErr:     tc.publishJobsErr,
				notifyPublisherErr: tc.notifyPublisherErr,
			}
			core, logs := observer.New(zap.DebugLevel)
			svc := New(s, m, tc.opt, zap.New(core))

			err := svc.Publish(context.TODO())

			a.Equal(tc.err, err)

			if tc.err != nil {
				return
			}

			l := make([]map[string]interface{}, 0)
			for _, i := range logs.All() {
				lm := i.ContextMap()
				lm["msg"] = i.Message
				l = append(l, lm)
			}
			a.ElementsMatch(tc.expectedLogs, l)
		})
	}
}

func getTestJobs() map[string][]storage.Job {
	return map[string][]storage.Job{
		"ec2": {
			{
				ID:     1,
				Host:   "ec2",
				Status: queuedStatus,
				Content: storage.JobContent{
					ID:         1,
					Owner:      "owner_1",
					Repository: "repo_1",
					Labels:     []string{"ec2", "ubuntu"},
				},
			},
			{
				ID:     2,
				Host:   "ec2",
				Status: inProgressStatus,
				Content: storage.JobContent{
					ID:         2,
					Owner:      "owner_2",
					Repository: "repo_2",
					Labels:     []string{"ec2", "windows"},
				},
			},
			{
				ID:     3,
				Host:   "ec2",
				Status: queuedStatus,
				Content: storage.JobContent{
					ID:         3,
					Owner:      "owner_3",
					Repository: "repo_3",
					Labels:     []string{"ec2", "windows"},
				},
			},
		},
		"eks": {
			{
				ID:     4,
				Host:   "eks",
				Status: queuedStatus,
				Content: storage.JobContent{
					ID:         4,
					Owner:      "owner_4",
					Repository: "repo_4",
					Labels:     []string{"eks", "ubuntu"},
				},
			},
			{
				ID:     5,
				Host:   "eks",
				Status: completedStatus,
				Content: storage.JobContent{
					ID:         5,
					Owner:      "owner_5",
					Repository: "repo_5",
					Labels:     []string{"eks", "windows"},
				},
			},
		},
	}
}

type mockedStorage struct {
	sync.RWMutex
	updateJobsInput *storage.UpdateJobsInput
	jobs            map[string][]storage.Job
	getJobsErr      error
	updateJobsErr   error
}

func (m *mockedStorage) GetJobs(_ context.Context, input *storage.GetJobsInput) ([]storage.Job, error) {
	jobs := m.jobs[input.Host]

	res := make([]storage.Job, 0)
	for i, v := range jobs {
		if inSlice(v.Status, input.Statuses) && int32(i) < input.Limit {
			res = append(res, v)
		}
	}

	return res, m.getJobsErr
}

func (m *mockedStorage) UpdateJobs(_ context.Context, input *storage.UpdateJobsInput) error {
	m.Lock()
	defer m.Unlock()

	m.updateJobsInput = input
	return m.updateJobsErr
}

type mockedMessenger struct {
	sync.RWMutex
	messages           []messenger.Message
	publishJobsErr     error
	notifyPublisherErr error
}

func (m *mockedMessenger) PublishJobs(_ context.Context, messages []messenger.Message) error {
	m.Lock()
	defer m.Unlock()

	m.messages = append(m.messages, messages...)
	return m.publishJobsErr
}

func (m *mockedMessenger) NotifyPublisher(_ context.Context) error {
	return m.notifyPublisherErr
}

func inSlice(key string, s []string) bool {
	for _, i := range s {
		if key == i {
			return true
		}
	}

	return false
}
