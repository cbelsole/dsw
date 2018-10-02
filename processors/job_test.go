package processors_test

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cbelsole/dsw/processors"
	"github.com/cbelsole/dsw/types"
	"github.com/satori/go.uuid"
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
	return nil
}

func TestJobSupportsExponentialBackoffOnFailure(t *testing.T) {
	var invocationTimes []time.Time

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		invocationTimes = append(invocationTimes, now)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	alwaysFails := types.Job{
		ID:        uuid.NewV4(),
		URI:       ts.URL,
		ExecuteAt: time.Now().Add(-10 * time.Second),
	}

	store := fakeJobStore{
		jobs: []*types.Job{&alwaysFails},
	}

	processor := processors.Job{
		Store:      store,
		WorkerNum:  1,
		MaxRetries: 5,
	}

	processor.Start()

	<-time.After(1 * time.Minute)

	expectedInvocationTimes := processor.MaxRetries + 1
	actualInvocationTimes := len(invocationTimes)

	if actualInvocationTimes != expectedInvocationTimes {
		t.Fatalf("%v should have been invoked %v times only saw %v invocations", ts.URL, expectedInvocationTimes, actualInvocationTimes)
	}

	const pollingInterval = 6 * time.Second

	for i := range invocationTimes {
		if i < len(invocationTimes)-1 {
			actualInterval := invocationTimes[i+1].Sub(invocationTimes[i]).Round(time.Second)
			expectedInterval := time.Duration(math.Exp2(float64(i)))*time.Second + pollingInterval
			delta := expectedInterval - actualInterval

			if delta > pollingInterval {
				t.Logf("All invocations: %v\n", invocationTimes)
				t.Errorf("%v backoff interval difference should have been <= 6 seconds between invocation %v and %v but was %v",
					ts.URL, i, i+1, delta)
			}
		}
	}
}
