package jobs

type hydratorJob struct{}

func NewHydratorJob() Job {
	return &hydratorJob{}
}

func (*hydratorJob) Start() {}

func (*hydratorJob) Stop() error {
	return nil
}
