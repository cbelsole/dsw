package types

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// Job contains the information needed to execute a job
type Job struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ErrorURI  *string   `json:"error_uri" db:"error_uri"`
	ExecuteAt time.Time `json:"execute_at" db:"execute_at"`
	Payload   []byte    `json:"payload" db:"payload"`
	URI       string    `json:"uri" db:"uri"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
