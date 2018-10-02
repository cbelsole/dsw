package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/cbelsole/dsw/processors"

	"github.com/db-journey/migrate"
	_ "github.com/db-journey/postgresql-driver"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"github.com/cbelsole/dsw/db"
	"github.com/cbelsole/dsw/handlers"
)

const maxRetries = 3

func main() {
	if err := runMigrations(); err != nil {
		log.Printf("failed to run migrations: %s\n", err)
		os.Exit(1)
	} else {
		log.Println("migrations completed successfully")
	}

	d, err := sql.Open("postgres", fmt.Sprintf("%s?sslmode=disable&timezone=UTC", os.Getenv("POSTGRES_URL")))
	if err != nil {
		log.Printf("unable to open db: %s\n", err)
		os.Exit(1)
	}

	database := db.NewDB(d)
	processor := processors.Job{Store: database, WorkerNum: 3, MaxRetries: 3}
	if err := processor.Start(); err != nil {
		log.Fatal(err)
	}

	h := handlers.Handler{DB: database, Job: processor}
	r := mux.NewRouter()
	r.Use(handlers.RecoveryMiddleware, handlers.LoggingMiddleware)

	// health
	r.HandleFunc("/", h.HealthHandler).Methods("GET")
	r.HandleFunc("/health", h.HealthHandler).Methods("GET")

	// jobs
	r.HandleFunc("/jobs", h.CreateJob).Methods("POST")
	r.HandleFunc("/jobs", h.ListJobs).Methods("GET")

	server := http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Println("server listening at 8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	graceful(&server, 5*time.Second)
}

func graceful(hs *http.Server, timeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := hs.Shutdown(ctx); err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Println("Server stopped")
	}
}

func runMigrations() error {
	dbURL := fmt.Sprintf("%s?sslmode=disable&timezone=UTC", os.Getenv("POSTGRES_URL"))
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	retries := 1
	var migrator *migrate.Handle

	for retries <= maxRetries {
		migrator, err = migrate.Open(dbURL, path.Join(dir, "migrations"))
		if err != nil {
			log.Printf("migrator failed to open (try %d out of 3)", retries)
			if retries == maxRetries {
				return err
			}
			retries++
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	defer migrator.Close()

	return migrator.Up(context.Background())
}
