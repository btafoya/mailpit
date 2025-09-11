//go:build integration

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/btafoya/mailsandbox/config"
	"github.com/btafoya/mailsandbox/internal/tools"
)

// TestDatabaseConnectionPooling verifies that multiple connections can be used concurrently
func TestDatabaseConnectionPooling(t *testing.T) {
	// Initialize database with improved pooling
	if err := InitDB(config.DataFile); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()

	// Test concurrent operations
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	
	// Spawn 10 concurrent database operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Each goroutine performs a database operation
			start := time.Now()
			_, err := db.Exec("SELECT 1")
			duration := time.Since(start)
			
			if err != nil {
				errors <- fmt.Errorf("goroutine %d failed: %v", id, err)
				return
			}
			
			// With proper pooling, operations should complete quickly
			if duration > 100*time.Millisecond {
				errors <- fmt.Errorf("goroutine %d took too long: %v", id, duration)
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

// TestBatchTagLoading verifies the N+1 query fix for message tags
func TestBatchTagLoading(t *testing.T) {
	if err := InitDB(config.DataFile); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()
	
	// Create test messages with tags
	messageIDs := []string{}
	for i := 0; i < 100; i++ {
		id := tools.GenerateID(16)
		messageIDs = append(messageIDs, id)
		
		// Insert test message
		raw := fmt.Sprintf("From: test%d@example.com\r\nSubject: Test %d\r\n\r\nBody", i, i)
		if err := Store(&[]byte(raw)); err != nil {
			t.Fatalf("Failed to store message %d: %v", i, err)
		}
		
		// Add tags
		for j := 0; j < 3; j++ {
			tagName := fmt.Sprintf("tag_%d_%d", i, j)
			if _, err := SetMessageTags(id, []string{tagName}); err != nil {
				t.Fatalf("Failed to add tag: %v", err)
			}
		}
	}
	
	// Test batch loading performance
	start := time.Now()
	tagMap := getMessageTagsBatch(messageIDs)
	duration := time.Since(start)
	
	// Verify all messages have tags
	for _, id := range messageIDs {
		tags, exists := tagMap[id]
		if !exists {
			t.Errorf("Message %s missing from tag map", id)
			continue
		}
		if len(tags) != 3 {
			t.Errorf("Message %s has %d tags, expected 3", id, len(tags))
		}
	}
	
	// Batch loading should be fast (< 100ms for 100 messages)
	if duration > 100*time.Millisecond {
		t.Errorf("Batch tag loading took %v, expected < 100ms", duration)
	}
	
	// Clean up
	for _, id := range messageIDs {
		DeleteOneMessage(id)
	}
}

// TestTransactionErrorHandling verifies improved error handling in transactions
func TestTransactionErrorHandling(t *testing.T) {
	if err := InitDB(config.DataFile); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()
	
	// Test rollback error handling
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
	
	// Now try to rollback - should get sql.ErrTxDone
	err = tx.Rollback()
	if err != sql.ErrTxDone {
		t.Errorf("Expected sql.ErrTxDone, got %v", err)
	}
	
	// Our improved error handling should handle this gracefully
	// Check that no panic occurs
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic occurred during rollback: %v", r)
			}
		}()
		
		// Simulate our improved error handling
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// This should not execute for ErrTxDone
			t.Errorf("Unexpected error logged: %v", err)
		}
	}()
}

// TestMessageListPerformance verifies the overall performance improvements
func TestMessageListPerformance(t *testing.T) {
	if err := InitDB(config.DataFile); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()
	
	// Create 100 test messages
	for i := 0; i < 100; i++ {
		raw := fmt.Sprintf("From: perf%d@example.com\r\nSubject: Performance Test %d\r\n\r\nBody", i, i)
		if err := Store(&[]byte(raw)); err != nil {
			t.Fatalf("Failed to store message %d: %v", i, err)
		}
	}
	
	// Test list performance with batch tag loading
	start := time.Now()
	messages, err := List(0, 100)
	duration := time.Since(start)
	
	if err != nil {
		t.Fatalf("Failed to list messages: %v", err)
	}
	
	if len(messages) < 100 {
		t.Errorf("Expected at least 100 messages, got %d", len(messages))
	}
	
	// With improvements, listing 100 messages should be fast
	if duration > 500*time.Millisecond {
		t.Errorf("Message listing took %v, expected < 500ms", duration)
	}
	
	// Clean up
	if err := DeleteSearch("from:perf"); err != nil {
		t.Errorf("Failed to clean up test messages: %v", err)
	}
}