package processors

import (
	"github.com/cbelsole/dsw/db"
	"github.com/cbelsole/dsw/types"
)

// Job is a processor responsible for enqueuing, running, and completing jobs
type Job struct {
	DB *db.DB
}

// Start adds a job to the pool
func (j *Job) New(db *db.DB) *Job {
	return &Job{DB: db}
}

// Start adds a job to the pool
func (j *Job) Start() error {
	return nil
}

// Enqueue adds a job to the pool
func (j *Job) Enqueue(job *types.Job) error {
	return nil
}
