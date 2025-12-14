package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/swarm-indexer/swarm-indexer/internal/chunker"
	"github.com/swarm-indexer/swarm-indexer/internal/config"
	"github.com/swarm-indexer/swarm-indexer/internal/detector"
	"github.com/swarm-indexer/swarm-indexer/internal/embeddings"
	"github.com/swarm-indexer/swarm-indexer/internal/metadata"
	"github.com/swarm-indexer/swarm-indexer/internal/secrets"
	"github.com/swarm-indexer/swarm-indexer/internal/walker"
)

// TestNewIndexer tests that an indexer can be created with configuration
func TestNewIndexer(t *testing.T) {
	cfg := &config.Config{
		Workers:   4,
		BatchSize: 50,
	}

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}
	if idx == nil {
		t.Fatal("NewIndexer returned nil")
	}
	if idx.workers != 4 {
		t.Errorf("expected workers=4, got %d", idx.workers)
	}
	if idx.batchSize != 50 {
		t.Errorf("expected batchSize=50, got %d", idx.batchSize)
	}
}

// TestIndexerWorkerPoolRespectsConcurrency tests that worker pool respects worker count
func TestIndexerWorkerPoolRespectsConcurrency(t *testing.T) {
	// Setup test directory with multiple files
	tempDir := t.TempDir()
	for i := 0; i < 20; i++ {
		path := filepath.Join(tempDir, "file"+string(rune('a'+i))+".go")
		if err := os.WriteFile(path, []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Workers:   3, // Low worker count to test concurrency
		BatchSize: 10,
	}

	// Track maximum concurrent workers
	var currentWorkers atomic.Int32
	var maxConcurrent atomic.Int32

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}

	// Set up tracking callback
	idx.SetProcessingCallback(func(start bool) {
		if start {
			current := currentWorkers.Add(1)
			for {
				max := maxConcurrent.Load()
				if current <= max || maxConcurrent.CompareAndSwap(max, current) {
					break
				}
			}
		} else {
			currentWorkers.Add(-1)
		}
	})

	ctx := context.Background()
	err = idx.IndexPaths(ctx, []string{tempDir})
	if err != nil {
		t.Fatalf("IndexPaths failed: %v", err)
	}

	// Verify max concurrent didn't exceed worker count
	if maxConcurrent.Load() > int32(cfg.Workers) {
		t.Errorf("max concurrent workers (%d) exceeded configured workers (%d)",
			maxConcurrent.Load(), cfg.Workers)
	}
}

// TestIndexerFullPipeline tests the full indexing pipeline
func TestIndexerFullPipeline(t *testing.T) {
	// Setup test directory with files
	tempDir := t.TempDir()
	testFiles := map[string]string{
		"main.go":    "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
		"lib.go":     "package main\n\nfunc helper() int {\n\treturn 42\n}\n",
		"readme.txt": "This is a readme file.\n",
	}

	for name, content := range testFiles {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Workers:   2,
		BatchSize: 10,
	}

	// Create mock to track pipeline stages
	tracker := &PipelineTracker{}

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}
	idx.SetPipelineTracker(tracker)

	ctx := context.Background()
	err = idx.IndexPaths(ctx, []string{tempDir})
	if err != nil {
		t.Fatalf("IndexPaths failed: %v", err)
	}

	// Verify all pipeline stages were executed
	if tracker.WalkCount == 0 {
		t.Error("walk stage not executed")
	}
	if tracker.SecretScanCount == 0 {
		t.Error("secret scan stage not executed")
	}
	if tracker.ChunkCount == 0 {
		t.Error("chunk stage not executed")
	}
	if tracker.EmbedCount == 0 {
		t.Error("embed stage not executed")
	}
	if tracker.UpsertCount == 0 {
		t.Error("upsert stage not executed")
	}
}

// TestIndexerBatchProcessing tests that batching is respected
func TestIndexerBatchProcessing(t *testing.T) {
	// Setup test directory with files
	tempDir := t.TempDir()
	for i := 0; i < 25; i++ {
		path := filepath.Join(tempDir, "file"+string(rune('a'+i))+".go")
		if err := os.WriteFile(path, []byte("package main\nfunc f() {}\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Workers:   4,
		BatchSize: 5, // Small batch size
	}

	var upsertBatchSizes []int
	var mu sync.Mutex

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}

	idx.SetUpsertCallback(func(batchSize int) {
		mu.Lock()
		upsertBatchSizes = append(upsertBatchSizes, batchSize)
		mu.Unlock()
	})

	ctx := context.Background()
	err = idx.IndexPaths(ctx, []string{tempDir})
	if err != nil {
		t.Fatalf("IndexPaths failed: %v", err)
	}

	// Verify batches were used
	if len(upsertBatchSizes) == 0 {
		t.Error("no upsert batches recorded")
	}

	// Verify batch sizes don't exceed configured maximum
	for i, size := range upsertBatchSizes {
		if size > cfg.BatchSize {
			t.Errorf("batch %d size (%d) exceeded configured batch size (%d)",
				i, size, cfg.BatchSize)
		}
	}
}

// TestIndexerIncrementalSkipsUnchanged tests incremental indexing
func TestIndexerIncrementalSkipsUnchanged(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Workers:   2,
		BatchSize: 10,
	}

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}

	tracker := &PipelineTracker{}
	idx.SetPipelineTracker(tracker)

	ctx := context.Background()

	// First index run
	err = idx.IndexPaths(ctx, []string{tempDir})
	if err != nil {
		t.Fatalf("first IndexPaths failed: %v", err)
	}

	firstRunUpserts := tracker.UpsertCount

	// Reset tracker
	tracker.Reset()

	// Second index run without changes
	err = idx.IndexPaths(ctx, []string{tempDir})
	if err != nil {
		t.Fatalf("second IndexPaths failed: %v", err)
	}

	// Verify second run skipped processing (incremental)
	if tracker.UpsertCount >= firstRunUpserts && firstRunUpserts > 0 {
		t.Errorf("expected incremental to skip unchanged, but got %d upserts (first run had %d)",
			tracker.UpsertCount, firstRunUpserts)
	}
}

// TestIndexerErrorHandlingContinuesOnFailure tests graceful error handling
func TestIndexerErrorHandlingContinuesOnFailure(t *testing.T) {
	// Setup test directory with files
	tempDir := t.TempDir()
	goodFiles := []string{"good1.go", "good2.go", "good3.go"}
	for _, name := range goodFiles {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Workers:   2,
		BatchSize: 10,
	}

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}

	// Set up a callback that fails for specific files
	var processedFiles []string
	var mu sync.Mutex
	idx.SetFileProcessorWithError(func(path string) error {
		mu.Lock()
		processedFiles = append(processedFiles, filepath.Base(path))
		mu.Unlock()

		// Fail for one specific file
		if filepath.Base(path) == "good2.go" {
			return &testError{"simulated failure"}
		}
		return nil
	})

	ctx := context.Background()
	err = idx.IndexPaths(ctx, []string{tempDir})

	// The indexer should continue despite errors
	mu.Lock()
	count := len(processedFiles)
	mu.Unlock()

	// All files should be attempted
	if count < len(goodFiles) {
		t.Errorf("expected all %d files to be processed, but only %d were",
			len(goodFiles), count)
	}
}

// TestIndexerProgressLogging tests that progress is logged
func TestIndexerProgressLogging(t *testing.T) {
	tempDir := t.TempDir()
	for i := 0; i < 10; i++ {
		path := filepath.Join(tempDir, "file"+string(rune('a'+i))+".go")
		if err := os.WriteFile(path, []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Workers:   2,
		BatchSize: 5,
	}

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}

	var progressUpdates []int
	var mu sync.Mutex
	idx.SetProgressCallback(func(processed int) {
		mu.Lock()
		progressUpdates = append(progressUpdates, processed)
		mu.Unlock()
	})

	ctx := context.Background()
	err = idx.IndexPaths(ctx, []string{tempDir})
	if err != nil {
		t.Fatalf("IndexPaths failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(progressUpdates) == 0 {
		t.Error("expected progress updates to be logged")
	}
}

// TestIndexerMetadataUpdate tests that metadata is updated after successful index
func TestIndexerMetadataUpdate(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Workers:   2,
		BatchSize: 10,
	}

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}

	ctx := context.Background()
	err = idx.IndexPaths(ctx, []string{tempDir})
	if err != nil {
		t.Fatalf("IndexPaths failed: %v", err)
	}

	// Check that metadata file was created
	metadataPath := filepath.Join(tempDir, metadata.MetadataFileName)
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("metadata file was not created")
	}

	// Load and verify metadata
	mgr := metadata.New()
	meta, err := mgr.Load(tempDir)
	if err != nil {
		t.Fatalf("failed to load metadata: %v", err)
	}
	if meta == nil {
		t.Fatal("metadata is nil")
	}
	if meta.ContentHash == "" {
		t.Error("metadata content hash is empty")
	}
	if meta.LastIndexed == 0 {
		t.Error("metadata last_indexed is not set")
	}
}

// TestIndexerContextCancellation tests that indexer respects context cancellation
func TestIndexerContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	// Create many files to ensure cancellation has time to happen
	for i := 0; i < 100; i++ {
		path := filepath.Join(tempDir, fmt.Sprintf("file%03d.go", i))
		if err := os.WriteFile(path, []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Workers:   1, // Single worker to make cancellation more predictable
		BatchSize: 10,
	}

	idx, err := NewIndexer(cfg)
	if err != nil {
		t.Fatalf("NewIndexer failed: %v", err)
	}

	// Add a slow processing callback
	idx.SetProcessingCallback(func(start bool) {
		if start {
			time.Sleep(10 * time.Millisecond)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err = idx.IndexPaths(ctx, []string{tempDir})

	// Should get context cancellation error
	if err == nil {
		t.Error("expected error from cancelled context")
	}
	if err != context.Canceled {
		t.Logf("got error: %v (type: %T)", err, err)
		// Accept any error since cancellation timing can vary
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Compile-time interface checks to ensure dependencies exist
var (
	_ = walker.New(nil)
	_ = detector.New()
	_ = secrets.New()
	_ = chunker.New(0)
	_ = embeddings.New("", "", 0)
	_ = metadata.New()
	_ = NewTypesenseClient("", "", "")
)
