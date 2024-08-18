package jobs

type Job interface {
	Start()
	Stop() error
}
