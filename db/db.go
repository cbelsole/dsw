package db

import (
	"database/sql"
	"time"

	"github.com/cbelsole/dsw/types"
	"github.com/helloeave/json"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

type (
	DB struct {
		DB *sqlx.DB
	}
	job struct {
		ID        uuid.UUID       `db:"id"`
		Errors    json.RawMessage `db:"errors"`
		ErrorURI  *string         `db:"error_uri"`
		ExecuteAt time.Time       `db:"execute_at"`
		Payload   json.RawMessage `db:"payload"`
		Sent      bool            `db:"sent"`
		Try       int             `db:"try"`
		URI       string          `db:"uri"`
		CreatedAt time.Time       `db:"created_at"`
		UpdatedAt time.Time       `db:"updated_at"`
	}
)

func toDBJob(j *types.Job) (*job, error) {
	errors, err := json.MarshalSafeCollections(j.Errors)
	if err != nil {
		return nil, err
	}

	payload, err := json.MarshalSafeCollections(j.Payload)
	if err != nil {
		return nil, err
	}

	return &job{
		ID:        j.ID,
		Errors:    json.RawMessage(errors),
		ErrorURI:  j.ErrorURI,
		ExecuteAt: j.ExecuteAt,
		Payload:   json.RawMessage(payload),
		Sent:      j.Sent,
		Try:       j.Try,
		URI:       j.URI,
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
	}, nil
}

func (j *job) toJob() (*types.Job, error) {
	var errors []string
	if err := json.Unmarshal(j.Errors, &errors); err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(j.Payload, &payload); err != nil {
		return nil, err
	}

	return &types.Job{
		ID:        j.ID,
		Errors:    errors,
		ErrorURI:  j.ErrorURI,
		ExecuteAt: j.ExecuteAt,
		Payload:   payload,
		Sent:      j.Sent,
		Try:       j.Try,
		URI:       j.URI,
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
	}, nil
}

func NewDB(d *sql.DB) *DB {
	return &DB{DB: sqlx.NewDb(d, "postgres")}
}

// Ping is a tiny method to make sure the db is alive
func (db *DB) Ping() error {
	_, err := db.DB.Exec("SELECT 1")
	return err
}

// CreateJob is a tiny method to make sure the db is alive
func (db *DB) CreateJob(job *types.Job) error {
	dbJob, err := toDBJob(job)
	if err != nil {
		return err
	}

	rows, err := db.DB.NamedQuery(
		"INSERT into jobs (uri,error_uri,payload,execute_at) VALUES (:uri,:error_uri,:payload,:execute_at) RETURNING *",
		dbJob,
	)

	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.StructScan(dbJob)
	}
	j, err := dbJob.toJob()
	*job = *j

	return nil
}

func (db *DB) UpdateJob(job *types.Job) error {
	job.UpdatedAt = time.Now()
	dbJob, err := toDBJob(job)
	if err != nil {
		return err
	}

	_, err = db.DB.NamedExec("UPDATE jobs set errors = :errors, sent = :sent, try = :try, updated_at = :updated_at where id = :id", dbJob)
	return err
}

func (db *DB) GetJobs() ([]*types.Job, error) {
	var dbJobs []*job
	if err := db.DB.Select(&dbJobs, "SELECT * from jobs"); err != nil {
		return nil, err
	}

	jobs := make([]*types.Job, 0, len(dbJobs))

	for _, dbJob := range dbJobs {
		job, err := dbJob.toJob()
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetPendingJobs gets jobs where try > -1 and try < 4 and sent is false
func (db *DB) GetPendingJobs() ([]*types.Job, error) {
	var dbJobs []*job
	if err := db.DB.Select(&dbJobs, "SELECT * from jobs where try > -1 AND try < 3 AND sent is false"); err != nil {
		return nil, err
	}

	jobs := make([]*types.Job, 0, len(dbJobs))

	for _, dbJob := range dbJobs {
		job, err := dbJob.toJob()
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}
