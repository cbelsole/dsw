package db

import (
	"database/sql"

	"github.com/cbelsole/dsw/types"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	DB *sqlx.DB
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
	rows, err := db.DB.NamedQuery(
		"INSERT into jobs (uri,error_uri,payload,execute_at) VALUES (:uri,:error_uri,:payload,:execute_at) RETURNING *",
		job,
	)

	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.StructScan(job)
	}

	return nil
}

func (db *DB) GetJobs() ([]*types.Job, error) {
	var jobs []*types.Job
	if err := db.DB.Select(&jobs, "SELECT * from jobs"); err != nil {
		return nil, err
	}

	return jobs, nil
}
