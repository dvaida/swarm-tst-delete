package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/swarm-indexer/swarm-indexer/internal/chunker"
	"github.com/swarm-indexer/swarm-indexer/internal/config"
	"github.com/swarm-indexer/swarm-indexer/internal/detector"
	"github.com/swarm-indexer/swarm-indexer/internal/embeddings"
	"github.com/swarm-indexer/swarm-indexer/internal/metadata"
	"github.com/swarm-indexer/swarm-indexer/internal/secrets"
	"github.com/swarm-indexer/swarm-indexer/internal/walker"
)

// ProcessingCallback is called when processing starts/stops
type ProcessingCallback func(start bool)

// UpsertCallback is called when an upsert batch is performed
type UpsertCallback func(batchSize int)

// ProgressCallback is called to report progress
type ProgressCallback func(processed int)

// FileProcessorWithError allows testing error handling
type FileProcessorWithError func(path string) error

// fileJob represents a file to be processed
type fileJob struct {
	path        string
	relPath     string
	projectPath string
}

// PipelineTracker tracks pipeline stage executions for testing
type PipelineTracker struct {
	mu              sync.Mutex
	WalkCount       int
	SecretScanCount int
	ChunkCount      int
	EmbedCount      int
	UpsertCount     int
}

// RecordWalk increments walk count
func (p *PipelineTracker) RecordWalk() { p.mu.Lock(); p.WalkCount++; p.mu.Unlock() }

// RecordSecretScan increments secret scan count
func (p *PipelineTracker) RecordSecretScan() { p.mu.Lock(); p.SecretScanCount++; p.mu.Unlock() }

// RecordChunk increments chunk count
func (p *PipelineTracker) RecordChunk() { p.mu.Lock(); p.ChunkCount++; p.mu.Unlock() }

// RecordEmbed increments embed count
func (p *PipelineTracker) RecordEmbed() { p.mu.Lock(); p.EmbedCount++; p.mu.Unlock() }

// RecordUpsert increments upsert count
func (p *PipelineTracker) RecordUpsert() { p.mu.Lock(); p.UpsertCount++; p.mu.Unlock() }

// Reset resets all counters
func (p *PipelineTracker) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.WalkCount = 0
	p.SecretScanCount = 0
	p.ChunkCount = 0
	p.EmbedCount = 0
	p.UpsertCount = 0
}

// Indexer orchestrates the full indexing pipeline with worker pool
type Indexer struct {
	walker    *walker.Walker
	detector  *detector.Detector
	secrets   *secrets.Scanner
	chunker   *chunker.Chunker
	gemini    *embeddings.GeminiClient
	typesense *TypesenseClient
	metadata  *metadata.Manager
	workers   int
	batchSize int

	// Testing callbacks
	processingCallback   ProcessingCallback
	upsertCallback       UpsertCallback
	progressCallback     ProgressCallback
	fileProcessorWithErr FileProcessorWithError
	pipelineTracker      *PipelineTracker
}

// NewIndexer creates a new Indexer with the given configuration
func NewIndexer(cfg *config.Config) (*Indexer, error) {
	workers := cfg.Workers
	if workers <= 0 {
		workers = 8
	}

	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	return &Indexer{
		walker:    walker.New(cfg.SkipFiles),
		detector:  detector.New(),
		secrets:   secrets.New(),
		chunker:   chunker.New(0),
		gemini:    embeddings.New(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiRateLimit),
		typesense: NewTypesenseClient(cfg.TypesenseURL, cfg.TypesenseAPIKey, cfg.TypesenseCollection),
		metadata:  metadata.New(),
		workers:   workers,
		batchSize: batchSize,
	}, nil
}

// IndexPaths indexes all files in the given paths
func (i *Indexer) IndexPaths(ctx context.Context, paths []string) error {
	for _, path := range paths {
		if err := i.indexPath(ctx, path); err != nil {
			if err == context.Canceled {
				return err
			}
			log.Printf("error indexing path %s: %v", path, err)
		}
	}
	return nil
}

func (i *Indexer) indexPath(ctx context.Context, path string) error {
	// Record walk if tracking
	if i.pipelineTracker != nil {
		i.pipelineTracker.RecordWalk()
	}

	// Check for incremental indexing
	existingMeta, err := i.metadata.Load(path)
	if err != nil {
		log.Printf("warning: failed to load metadata: %v", err)
	}

	hasChanged, newHash, err := i.metadata.HasChanged(path, existingMeta)
	if err != nil {
		log.Printf("warning: failed to compute content hash: %v", err)
		hasChanged = true // Index anyway if we can't check
	}

	if !hasChanged {
		log.Printf("skipping unchanged path: %s", path)
		return nil
	}

	// Walk directory
	files, err := i.walker.Walk(ctx, path)
	if err != nil {
		return fmt.Errorf("walk failed: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	// Create job channel and result tracking
	jobs := make(chan fileJob, len(files))
	var wg sync.WaitGroup
	var processedCount atomic.Int32
	var chunkBatch []IndexedChunk
	var batchMu sync.Mutex
	var errors []error
	var errorsMu sync.Mutex

	// Start workers
	for w := 0; w < i.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Signal processing start
				if i.processingCallback != nil {
					i.processingCallback(true)
				}

				chunks, err := i.processFile(ctx, job)
				if err != nil {
					errorsMu.Lock()
					errors = append(errors, fmt.Errorf("processing %s: %w", job.path, err))
					errorsMu.Unlock()
					log.Printf("error processing %s: %v", job.path, err)
				}

				// Signal processing end
				if i.processingCallback != nil {
					i.processingCallback(false)
				}

				// Add to batch
				if len(chunks) > 0 {
					batchMu.Lock()
					chunkBatch = append(chunkBatch, chunks...)

					// Flush batch if full
					if len(chunkBatch) >= i.batchSize {
						batch := make([]IndexedChunk, len(chunkBatch))
						copy(batch, chunkBatch)
						chunkBatch = chunkBatch[:0]
						batchMu.Unlock()

						i.flushBatch(ctx, batch)
					} else {
						batchMu.Unlock()
					}
				}

				// Update progress
				count := processedCount.Add(1)
				if i.progressCallback != nil {
					i.progressCallback(int(count))
				}

				// Log progress every N files
				if count%10 == 0 {
					log.Printf("progress: %d/%d files processed", count, len(files))
				}
			}
		}()
	}

	// Send jobs
	for _, f := range files {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return ctx.Err()
		case jobs <- fileJob{path: f.Path, relPath: f.RelPath, projectPath: path}:
		}
	}
	close(jobs)

	// Wait for workers to finish
	wg.Wait()

	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Flush remaining batch
	batchMu.Lock()
	if len(chunkBatch) > 0 {
		batch := make([]IndexedChunk, len(chunkBatch))
		copy(batch, chunkBatch)
		batchMu.Unlock()
		i.flushBatch(ctx, batch)
	} else {
		batchMu.Unlock()
	}

	// Update metadata
	meta := &metadata.Metadata{
		LastIndexed: time.Now().Unix(),
		FileCount:   len(files),
		ContentHash: newHash,
		ProjectType: "unknown",
		Languages:   []string{},
	}

	if err := i.metadata.Save(path, meta); err != nil {
		log.Printf("warning: failed to save metadata: %v", err)
	}

	return nil
}

func (i *Indexer) processFile(ctx context.Context, job fileJob) ([]IndexedChunk, error) {
	// Custom error processor for testing
	if i.fileProcessorWithErr != nil {
		if err := i.fileProcessorWithErr(job.path); err != nil {
			return nil, err
		}
	}

	// Check if file should be skipped (Type A secrets)
	if i.pipelineTracker != nil {
		i.pipelineTracker.RecordSecretScan()
	}

	scanResult, err := i.secrets.ScanFile(job.path)
	if err != nil {
		return nil, fmt.Errorf("secret scan failed: %w", err)
	}
	if scanResult.ShouldSkip {
		return nil, nil
	}

	// Read file content
	content, err := i.walker.ReadFile(job.path)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	// Scan content for inline secrets and redact
	contentScanResult, err := i.secrets.ScanContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("content scan failed: %w", err)
	}

	redactedContent := i.secrets.Redact(string(content), contentScanResult.Findings)

	// Detect language
	language := i.detector.DetectLanguage(job.path)

	// Chunk content
	if i.pipelineTracker != nil {
		i.pipelineTracker.RecordChunk()
	}

	chunks, err := i.chunker.ChunkContent(redactedContent, language)
	if err != nil {
		return nil, fmt.Errorf("chunking failed: %w", err)
	}

	if len(chunks) == 0 {
		return nil, nil
	}

	// Generate embeddings
	if i.pipelineTracker != nil {
		i.pipelineTracker.RecordEmbed()
	}

	texts := make([]string, len(chunks))
	for j, chunk := range chunks {
		texts[j] = chunk.Content
	}

	embeds, err := i.gemini.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	// Create indexed chunks
	indexedChunks := make([]IndexedChunk, len(chunks))
	for j, chunk := range chunks {
		indexedChunks[j] = IndexedChunk{
			ID:          generateChunkID(job.path, chunk.StartLine),
			FilePath:    job.relPath,
			ProjectPath: job.projectPath,
			ProjectType: "unknown",
			Language:    language,
			ChunkType:   chunk.Type,
			Content:     chunk.Content,
			Embedding:   embeds[j],
			StartLine:   chunk.StartLine,
			EndLine:     chunk.EndLine,
			LastIndexed: time.Now().Unix(),
		}
	}

	return indexedChunks, nil
}

func (i *Indexer) flushBatch(ctx context.Context, batch []IndexedChunk) {
	if len(batch) == 0 {
		return
	}

	// Record upsert in pipeline tracker
	if i.pipelineTracker != nil {
		i.pipelineTracker.RecordUpsert()
	}

	// Callback for testing
	if i.upsertCallback != nil {
		i.upsertCallback(len(batch))
	}

	// Upsert to Typesense
	if err := i.typesense.UpsertBatch(ctx, batch); err != nil {
		log.Printf("warning: upsert batch failed: %v", err)
	}
}

func generateChunkID(path string, startLine int) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d", path, startLine)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// SetProcessingCallback sets a callback for processing events (for testing)
func (i *Indexer) SetProcessingCallback(cb ProcessingCallback) {
	i.processingCallback = cb
}

// SetUpsertCallback sets a callback for upsert events (for testing)
func (i *Indexer) SetUpsertCallback(cb UpsertCallback) {
	i.upsertCallback = cb
}

// SetProgressCallback sets a callback for progress events (for testing)
func (i *Indexer) SetProgressCallback(cb ProgressCallback) {
	i.progressCallback = cb
}

// SetFileProcessorWithError sets a custom file processor for testing error handling
func (i *Indexer) SetFileProcessorWithError(cb FileProcessorWithError) {
	i.fileProcessorWithErr = cb
}

// SetPipelineTracker sets a pipeline tracker for testing
func (i *Indexer) SetPipelineTracker(tracker *PipelineTracker) {
	i.pipelineTracker = tracker
}
