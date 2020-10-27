package worker

import (
	"github.com/robfig/cron/v3"
)

// type Worker interface {
// 	Run(ctx context.Context) error
// }

// IJob job的接口
type IJob interface {
	Start() error
	Run()
	Stop() error
}

type OnWork func() error

type BaseJob struct {
	Cron      *cron.Cron
	IsRunning bool
	OnWork    OnWork
}

func (job *BaseJob) Start() error {
	job.Cron.Start()
	return nil
}

func (job *BaseJob) Stop() error {
	job.Cron.Stop()
	return nil
}

func (job *BaseJob) Run() {
	if job.IsRunning {
		return
	}

	job.IsRunning = true

	job.OnWork()

	job.IsRunning = false
}
