package processors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/cbelsole/dsw/types"
)

type JobStore interface {
	UpdateJob(*types.Job) error
	GetJobs() ([]*types.Job, error)
	GetPendingJobs() ([]*types.Job, error)
	CreateJob(*types.Job) error
}

// Job is a processor responsible for enqueuing, running, and completing jobs
type Job struct {
	Store                 JobStore
	WorkerNum, MaxRetries int
}

var (
	jobs           sync.Map
	started        sync.Once
	jobQueue       = make(chan *types.Job, 100)
	results        = make(chan *types.Job, 100)
	processingJobs sync.Map
)

// Start adds a job to the pool
func (j *Job) Start() error {
	var err error
	started.Do(func() {
		var loadedJobs []*types.Job
		loadedJobs, err = j.Store.GetPendingJobs()
		if err != nil {
			return
		}

		for _, job := range loadedJobs {
			jobs.Store(job.ID.String(), job)
		}

		for w := 0; w < j.WorkerNum; w++ {
			go j.worker(w, jobQueue, results)
		}

		go func() {
			for range time.Tick(5 * time.Second) {
				for _, j := range j.getJobs() {
					jobQueue <- j
				}
			}
		}()

		go func() {
			for job := range results {
				if err := j.Store.UpdateJob(job); err != nil {
					log.Printf("error saving job %+v, error: %s\n", job, err)
				} else {
					log.Printf("processed job %+v\n", job)
				}

				// Remove completed jobs
				jobs.Store(job.ID.String(), job)
				processingJobs.Delete(job.ID.String())
			}
		}()
	})

	return err
}

// Enqueue adds a job to the pool
func (j *Job) Enqueue(job *types.Job) error {
	if err := j.Store.CreateJob(job); err != nil {
		return err
	}

	jobs.Store(job.ID.String(), job)

	return nil
}

func (j *Job) worker(id int, processing <-chan *types.Job, results chan<- *types.Job) {
	for job := range processing {
		log.Printf("starting job %+v\n", job)
		payload, err := json.Marshal(job.Payload)
		if err != nil {
			job.Errors = append(job.Errors, err.Error())
			results <- job
			continue
		}

		resp, err := http.Post(job.URI, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			job.Errors = append(job.Errors, err.Error())
			job.Try = -1
		} else {
			b, err := ioutil.ReadAll(resp.Body)
			fmt.Println("body: ", string(b))
			if err != nil {
				job.Errors = append(job.Errors, fmt.Sprintf("error reading body %s", err))
			}

			if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
				job.Sent = true
			} else if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
				job.Errors = append(job.Errors, fmt.Sprintf("URI returned 400: %s", string(b)))
				job.Try = -1
			} else if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
				job.Try++
			}

			resp.Body.Close()
		}

		log.Printf("finished job %+v\n", job)
		results <- job
	}
}

func (j *Job) getJobs() []*types.Job {
	var sendableJobs []*types.Job
	now := time.Now()

	jobs.Range(func(key, value interface{}) bool {
		job := value.(*types.Job)

		log.Printf("checking job %+v\n", job)
		// Remove completed jobs
		if job.Sent || job.Try == -1 || job.Try > j.MaxRetries {
			jobs.Delete(key)
			processingJobs.Delete(key)
			return true
		}

		// enqueue jobs that are not processing
		if _, found := processingJobs.Load(key); !found && !job.ExecuteAt.After(now) {
			sendableJobs = append(sendableJobs, job)
			processingJobs.Store(key, true)
		}

		return true
	})

	return sendableJobs
}
