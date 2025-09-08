package interfaces

import (
	"context"
	"time"
)

// Database interface defines common database operations
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error
	HealthCheck(ctx context.Context) error

	// Transaction support
	BeginTransaction(ctx context.Context) (Transaction, error)
	WithTransaction(ctx context.Context, fn func(Transaction) error) error

	// Basic operations
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)

	// Prepared statements
	Prepare(ctx context.Context, query string) (Stmt, error)

	// Connection pool info
	Stats() interface{}
}

// Transaction interface for database transactions
type Transaction interface {
	Commit() error
	Rollback() error
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
}

// Rows interface for query results
type Rows interface {
	Close() error
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

// Row interface for single row results
type Row interface {
	Scan(dest ...interface{}) error
}

// Result interface for execution results
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Stmt interface for prepared statements
type Stmt interface {
	Close() error
	Query(ctx context.Context, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, args ...interface{}) Row
	Exec(ctx context.Context, args ...interface{}) (Result, error)
}

// Cache interface defines common cache operations
type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Batch operations
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetMultiple(ctx context.Context, items map[string]interface{}, expiration time.Duration) error
	DeleteMultiple(ctx context.Context, keys []string) error

	// Advanced operations
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Pattern operations
	Keys(ctx context.Context, pattern string) ([]string, error)
	Flush(ctx context.Context) error

	// Connection management
	Close() error
	Ping(ctx context.Context) error
}

// MessageQueue interface defines common message queue operations
type MessageQueue interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	HealthCheck(ctx context.Context) error

	// Publishing
	Publish(ctx context.Context, topic string, message interface{}) error
	PublishWithKey(ctx context.Context, topic string, key string, message interface{}) error

	// Consuming
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	SubscribeWithGroup(ctx context.Context, topic, group string, handler MessageHandler) error

	// Queue management
	CreateTopic(ctx context.Context, topic string, partitions int) error
	DeleteTopic(ctx context.Context, topic string) error
	ListTopics(ctx context.Context) ([]string, error)
}

// MessageHandler defines the signature for message handlers
type MessageHandler func(ctx context.Context, message Message) error

// Message represents a message from the queue
type Message struct {
	Key       string
	Value     []byte
	Topic     string
	Partition int32
	Offset    int64
	Timestamp time.Time
	Headers   map[string]string
}

// Storage interface defines common storage operations
type Storage interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	HealthCheck(ctx context.Context) error

	// Bucket operations
	CreateBucket(ctx context.Context, bucketName string) error
	DeleteBucket(ctx context.Context, bucketName string) error
	ListBuckets(ctx context.Context) ([]string, error)
	BucketExists(ctx context.Context, bucketName string) (bool, error)

	// Object operations
	PutObject(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error
	GetObject(ctx context.Context, bucketName, objectName string) ([]byte, error)
	DeleteObject(ctx context.Context, bucketName, objectName string) error
	ObjectExists(ctx context.Context, bucketName, objectName string) (bool, error)

	// List operations
	ListObjects(ctx context.Context, bucketName, prefix string) ([]ObjectInfo, error)
	GetObjectInfo(ctx context.Context, bucketName, objectName string) (ObjectInfo, error)

	// Presigned URLs
	GetPresignedURL(ctx context.Context, bucketName, objectName string, expiration time.Duration) (string, error)
	GetPresignedUploadURL(ctx context.Context, bucketName, objectName string, expiration time.Duration) (string, error)
}

// ObjectInfo represents information about a stored object
type ObjectInfo struct {
	Name         string
	Size         int64
	LastModified time.Time
	ContentType  string
	ETag         string
}

// Search interface defines common search operations
type Search interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	HealthCheck(ctx context.Context) error

	// Index operations
	CreateIndex(ctx context.Context, indexName string, mapping interface{}) error
	DeleteIndex(ctx context.Context, indexName string) error
	IndexExists(ctx context.Context, indexName string) (bool, error)
	ListIndices(ctx context.Context) ([]string, error)

	// Document operations
	IndexDocument(ctx context.Context, indexName, documentID string, document interface{}) error
	GetDocument(ctx context.Context, indexName, documentID string) (interface{}, error)
	DeleteDocument(ctx context.Context, indexName, documentID string) error
	UpdateDocument(ctx context.Context, indexName, documentID string, document interface{}) error

	// Search operations
	Search(ctx context.Context, indexName string, query interface{}) (SearchResult, error)
	SearchMultiple(ctx context.Context, indices []string, query interface{}) (SearchResult, error)

	// Bulk operations
	BulkIndex(ctx context.Context, indexName string, documents []BulkDocument) error
	BulkDelete(ctx context.Context, indexName string, documentIDs []string) error
}

// BulkDocument represents a document for bulk operations
type BulkDocument struct {
	ID       string
	Document interface{}
}

// SearchResult represents search results
type SearchResult struct {
	Hits     []SearchHit
	Total    int64
	MaxScore float64
	Took     int64
}

// SearchHit represents a single search hit
type SearchHit struct {
	ID     string
	Score  float64
	Source interface{}
}
