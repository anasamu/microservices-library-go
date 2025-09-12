package types

import "time"

// StorageFeature represents a storage feature
type StorageFeature string

const (
	FeaturePresignedURLs StorageFeature = "presigned_urls"
	FeaturePublicURLs    StorageFeature = "public_urls"
	FeatureMultipart     StorageFeature = "multipart"
	FeatureVersioning    StorageFeature = "versioning"
	FeatureEncryption    StorageFeature = "encryption"
	FeatureLifecycle     StorageFeature = "lifecycle"
	FeatureCORS          StorageFeature = "cors"
)

// PutObjectRequest represents a request to put an object
type PutObjectRequest struct {
	Bucket      string                 `json:"bucket"`
	Key         string                 `json:"key"`
	Content     []byte                 `json:"content"`
	ContentType string                 `json:"content_type"`
	Metadata    map[string]string      `json:"metadata"`
	Size        int64                  `json:"size"`
	Options     map[string]interface{} `json:"options"`
}

// PutObjectResponse represents a response from putting an object
type PutObjectResponse struct {
	Key          string                 `json:"key"`
	ETag         string                 `json:"etag"`
	Size         int64                  `json:"size"`
	LastModified time.Time              `json:"last_modified"`
	Metadata     map[string]string      `json:"metadata"`
	VersionID    string                 `json:"version_id"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// GetObjectRequest represents a request to get an object
type GetObjectRequest struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	VersionID string `json:"version_id"`
	Range     *Range `json:"range"`
}

// Range represents a byte range for partial content requests
type Range struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// GetObjectResponse represents a response from getting an object
type GetObjectResponse struct {
	Content      []byte                 `json:"content"`
	Size         int64                  `json:"size"`
	ContentType  string                 `json:"content_type"`
	ETag         string                 `json:"etag"`
	LastModified time.Time              `json:"last_modified"`
	Metadata     map[string]string      `json:"metadata"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// DeleteObjectRequest represents a request to delete an object
type DeleteObjectRequest struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	VersionID string `json:"version_id"`
}

// DeleteObjectsRequest represents a request to delete multiple objects
type DeleteObjectsRequest struct {
	Bucket string   `json:"bucket"`
	Keys   []string `json:"keys"`
}

// DeleteObjectsResponse represents a response from deleting multiple objects
type DeleteObjectsResponse struct {
	Deleted []DeletedObject `json:"deleted"`
	Errors  []DeleteError   `json:"errors"`
}

// DeletedObject represents a successfully deleted object
type DeletedObject struct {
	Key string `json:"key"`
}

// DeleteError represents an error when deleting an object
type DeleteError struct {
	Key     string `json:"key"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ListObjectsRequest represents a request to list objects
type ListObjectsRequest struct {
	Bucket            string `json:"bucket"`
	Prefix            string `json:"prefix"`
	Delimiter         string `json:"delimiter"`
	MaxKeys           int32  `json:"max_keys"`
	Marker            string `json:"marker"`
	ContinuationToken string `json:"continuation_token"`
}

// ListObjectsResponse represents a response from listing objects
type ListObjectsResponse struct {
	Objects               []ObjectInfo           `json:"objects"`
	CommonPrefixes        []string               `json:"common_prefixes"`
	IsTruncated           bool                   `json:"is_truncated"`
	NextMarker            string                 `json:"next_marker"`
	NextContinuationToken string                 `json:"next_continuation_token"`
	ProviderData          map[string]interface{} `json:"provider_data"`
}

// ObjectInfo represents information about an object
type ObjectInfo struct {
	Key          string                 `json:"key"`
	Size         int64                  `json:"size"`
	ETag         string                 `json:"etag"`
	LastModified time.Time              `json:"last_modified"`
	ContentType  string                 `json:"content_type"`
	Metadata     map[string]string      `json:"metadata"`
	StorageClass string                 `json:"storage_class"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// ObjectExistsRequest represents a request to check if an object exists
type ObjectExistsRequest struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// CopyObjectRequest represents a request to copy an object
type CopyObjectRequest struct {
	SourceBucket string `json:"source_bucket"`
	SourceKey    string `json:"source_key"`
	DestBucket   string `json:"dest_bucket"`
	DestKey      string `json:"dest_key"`
}

// CopyObjectResponse represents a response from copying an object
type CopyObjectResponse struct {
	Key          string                 `json:"key"`
	ETag         string                 `json:"etag"`
	Size         int64                  `json:"size"`
	LastModified time.Time              `json:"last_modified"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// MoveObjectRequest represents a request to move an object
type MoveObjectRequest struct {
	SourceBucket string `json:"source_bucket"`
	SourceKey    string `json:"source_key"`
	DestBucket   string `json:"dest_bucket"`
	DestKey      string `json:"dest_key"`
}

// MoveObjectResponse represents a response from moving an object
type MoveObjectResponse struct {
	Key          string                 `json:"key"`
	ETag         string                 `json:"etag"`
	Size         int64                  `json:"size"`
	LastModified time.Time              `json:"last_modified"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// GetObjectInfoRequest represents a request to get object information
type GetObjectInfoRequest struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// PresignedURLRequest represents a request to generate a presigned URL
type PresignedURLRequest struct {
	Bucket    string        `json:"bucket"`
	Key       string        `json:"key"`
	Method    string        `json:"method"`
	ExpiresIn time.Duration `json:"expires_in"`
	Expires   time.Time     `json:"expires"`
}

// PublicURLRequest represents a request to generate a public URL
type PublicURLRequest struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// CreateBucketRequest represents a request to create a bucket
type CreateBucketRequest struct {
	Bucket string            `json:"bucket"`
	Region string            `json:"region"`
	ACL    string            `json:"acl"`
	Tags   map[string]string `json:"tags"`
}

// DeleteBucketRequest represents a request to delete a bucket
type DeleteBucketRequest struct {
	Bucket string `json:"bucket"`
}

// BucketExistsRequest represents a request to check if a bucket exists
type BucketExistsRequest struct {
	Bucket string `json:"bucket"`
}

// BucketInfo represents information about a bucket
type BucketInfo struct {
	Name         string                 `json:"name"`
	CreationDate time.Time              `json:"creation_date"`
	Region       string                 `json:"region"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// StorageStats represents storage statistics
type StorageStats struct {
	TotalObjects  int64                  `json:"total_objects"`
	TotalSize     int64                  `json:"total_size"`
	TotalBuckets  int64                  `json:"total_buckets"`
	RequestsCount int64                  `json:"requests_count"`
	ErrorsCount   int64                  `json:"errors_count"`
	LastRequest   time.Time              `json:"last_request"`
	ProviderData  map[string]interface{} `json:"provider_data"`
}

// ConnectionInfo represents connection information
type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	SSL      bool   `json:"ssl"`
}

// ProviderInfo represents provider information
type ProviderInfo struct {
	Name              string           `json:"name"`
	SupportedFeatures []StorageFeature `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo  `json:"connection_info"`
	IsConnected       bool             `json:"is_connected"`
}
