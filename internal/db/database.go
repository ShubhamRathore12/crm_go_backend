package db

import (
	"crm-backend/internal/config"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// QueuedWrite represents a write operation to be retried
type QueuedWrite struct {
	Query  string          `json:"query"`
	Params json.RawMessage `json:"params"`
	Retries int            `json:"retries"`
}

// DatabaseManager manages database connections
type DatabaseManager struct {
	primary    *sql.DB
	secondary  *sql.DB
	writeQueue chan *QueuedWrite
	mu         sync.RWMutex
	closed     bool
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(cfg *config.Config) (*DatabaseManager, error) {
	// Connect to primary database
	primary, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open primary database: %w", err)
	}

	// Set connection pool settings
	primary.SetMaxOpenConns(10)
	primary.SetMaxIdleConns(5)
	primary.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := primary.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping primary database: %w", err)
	}

	var secondary *sql.DB
	if cfg.SecondaryDatabaseURL != nil {
		secondary, err = sql.Open("postgres", *cfg.SecondaryDatabaseURL)
		if err != nil {
			log.Printf("Warning: failed to open secondary database: %v", err)
			secondary = nil
		} else {
			secondary.SetMaxOpenConns(5)
			secondary.SetMaxIdleConns(3)
			secondary.SetConnMaxLifetime(time.Hour)

			if err := secondary.Ping(); err != nil {
				log.Printf("Warning: failed to ping secondary database: %v", err)
				secondary = nil
			}
		}
	}

	db := &DatabaseManager{
		primary:    primary,
		secondary:  secondary,
		writeQueue: make(chan *QueuedWrite, 100),
	}

	// Start background worker for retry queue
	go db.processWriteQueue()

	return db, nil
}

// processWriteQueue processes queued writes in the background
func (db *DatabaseManager) processWriteQueue() {
	for item := range db.writeQueue {
		if db.closed {
			return
		}

		// Try to execute the query
		_, err := db.primary.Exec(item.Query)
		if err != nil {
			item.Retries++
			if item.Retries < 10 {
				log.Printf("Warning: Primary write failed, retrying (%d/10): %v", item.Retries, err)
				// Re-queue with delay
				time.Sleep(5 * time.Second)
				select {
				case db.writeQueue <- item:
				default:
					log.Printf("Error: Write queue full, dropping item")
				}
			} else {
				log.Printf("CRITICAL: Background write failed after 10 retries: %v", err)
			}
		} else {
			log.Println("Resilient write synced to primary")
		}
	}
}

// Primary returns the primary database connection
func (db *DatabaseManager) Primary() *sql.DB {
	return db.primary
}

// ReadPool returns a database suitable for reading
func (db *DatabaseManager) ReadPool() *sql.DB {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Test primary connection
	if err := db.primary.Ping(); err == nil {
		return db.primary
	}

	// Fallback to secondary if available
	if db.secondary != nil {
		log.Println("Primary unavailable, switching to secondary for READ")
		return db.secondary
	}

	return db.primary
}

// WritePool returns the database for writing (always primary)
func (db *DatabaseManager) WritePool() *sql.DB {
	return db.primary
}

// ResilientWrite queues a write operation that will be retried if it fails
func (db *DatabaseManager) ResilientWrite(query string) {
	_, err := db.primary.Exec(query)
	if err != nil {
		log.Printf("Direct write failed, queuing for retry: %v", err)
		select {
		case db.writeQueue <- &QueuedWrite{
			Query:   query,
			Params:  nil,
			Retries: 0,
		}:
		default:
			log.Println("Error: Write queue full")
		}
	}
}

// Close closes database connections
func (db *DatabaseManager) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.closed = true
	close(db.writeQueue)

	var err error
	if db.primary != nil {
		if closeErr := db.primary.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if db.secondary != nil {
		if closeErr := db.secondary.Close(); closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; also failed to close secondary: %v", err, closeErr)
			} else {
				err = closeErr
			}
		}
	}
	return err
}
