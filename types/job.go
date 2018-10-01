package types

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// Job contains the information needed to execute a job
type Job struct {
	ID        uuid.UUID              `json:"id"`
	Errors    []string               `json:"errors"`
	ErrorURI  *string                `json:"error_uri"`
	ExecuteAt time.Time              `json:"execute_at"`
	Payload   map[string]interface{} `json:"payload"`
	Sent      bool                   `json:"sent"`
	Try       int                    `json:"try"`
	URI       string                 `json:"uri"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}
