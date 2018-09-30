# dsw

dsw is a service to run jobs that makes a POST request to a URI at a specified time with an optional payload.

* On a 2xx response the job is logged as a success.
* On a 4xx response the request is not retried and logged as a failure.
* On a 5xx response the request is retried 3 times with exponential backoff and logged as a failure barring any successes on a retry.
* If the optional error_url parameter is passed it will be called with the payload on a failure following the above rules.

# Routes
## GET / and GET /health

```json
// Example response
// HTTP - 200

{
  "meta": {},
  "response": {
    "message": "I'm healthy"
  }
}
```
Error codes: 500

## GET /jobs
```json
// Example response
// HTTP - 200

{
    "meta": {},
    "response": [
        {
            "id": "7b596144-da13-4d93-ace7-4938bca2db76",
            "error_uri": "http://error.com/error",
            "execute_at": "2018-10-01T00:00:00Z",
            "payload": {
                "some": "data"
            },
            "uri": "http://test.com/test",
            "created_at": "2018-09-30T13:50:36.164374Z",
            "updated_at": "2018-09-30T13:50:36.164374Z"
        }
    ]
}
```
Error codes: 500

## POST /jobs
```json
// Example request
{
  "uri": "http://test.com/test",
  "execute_at": "2018-10-01T00:00:00Z",
  "payload": {
    "some": "data"
  },
  "error_uri": "http://error.com/error"
}
```

Optional parameters: error_uri, payload

```json
// Example response
// HTTP - 201

{
    "meta": {},
    "response": {
        "id": "7b596144-da13-4d93-ace7-4938bca2db76",
        "error_uri": "http://error.com/error",
        "execute_at": "2018-10-01T00:00:00Z",
        "payload": {
            "some": "data"
        },
        "uri": "http://test.com/test",
        "created_at": "2018-09-30T13:50:36.164374Z",
        "updated_at": "2018-09-30T13:50:36.164374Z"
    }
}
```
Error codes: 400,500

# Getting started

## Prerequisites

1. go `brew install go`
2. dep `brew install dep`

### .env file

Create this file in the project root

```
POSTGRES_URL=postgres://postgres:postgres@localhost:5432/postgres
ENVIRONMENT=development
```

## Running the app

`env $(cat .env | xargs) go run cmd/server/main.go`
