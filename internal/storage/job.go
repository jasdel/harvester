package storage

import (
	"github.com/jasdel/harvester/internal/types"
)

// public interface of a client Job. Hides implementation
// so clients cannot create instances of Jobs them selves
// since the properties are not exposed.
type Job interface {
	Id() types.JobId
	Status() (*types.JobStatus, error)
	Result() (*types.JobResult, error)
}

type job struct {
	id     types.JobId
	client *Client
}

func (j *job) Id() types.JobId {
	return j.id
}

func (j *job) Status() (*types.JobStatus, error) {
	return nil, nil
}

func (j *job) Result() (*types.JobResult, error) {
	return nil, nil
}
