package processors_test

import (
	"fmt"
	"github.com/cbelsole/dsw/processors"
	"github.com/cbelsole/dsw/types"
	"github.com/satori/go.uuid"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type fakeJobStore struct {
	jobs []*types.Job
}

func (f fakeJobStore) UpdateJob(*types.Job) error {
	return nil
}

func (f fakeJobStore) GetJobs() ([]*types.Job, error) {
	return f.jobs, nil
}

func (f fakeJobStore) GetPendingJobs() ([]*types.Job, error) {
	return f.jobs, nil
}

func (f fakeJobStore) CreateJob(*types.Job) error {
	return  nil
}

func TestJobSupportsRetriesOnFailure(t *testing.T) {
	var invocationTimes []time.Time

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		invocationTimes = append(invocationTimes, now)
		fmt.Fprintf(os.Stdout, "invoked at %v\n", now)
		fmt.Fprintf(os.Stdout, "all invocations %v\n", invocationTimes)
		w.WriteHeader(http.StatusInternalServerError)

	}))
	defer ts.Close()

	alwaysFails := types.Job{
		ID: uuid.NewV4(),
		URI: ts.URL,
		ExecuteAt: time.Now().Add(-10*time.Minute),
	}

	store := fakeJobStore{
		jobs: []*types.Job{&alwaysFails},
	}

	processor := processors.Job{
		Store: store,
		WorkerNum: 1,
		MaxRetries: 1,
	}

	processor.Start()

	<-time.After(15*time.Second)

	if len(invocationTimes) != 2 {
		t.Fatalf("%v should have been invoked %v times only saw %v invocations", ts.URL, 2, len(invocationTimes))
	}
}
