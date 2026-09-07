package documents

import (
	"context"
	"errors"
	"time"
)

const MaxDocumentSize int64 = 25 * 1024 * 1024

type Document struct {
	ID               string     `json:"id"`
	OrganizationID   string     `json:"organizationId"`
	ResourceType     string     `json:"resourceType"`
	ResourceID       string     `json:"resourceId"`
	Kind             string     `json:"kind"`
	FileName         string     `json:"fileName"`
	ContentType      string     `json:"contentType"`
	SizeBytes        int64      `json:"sizeBytes"`
	StorageKey       string     `json:"-"`
	Status           string     `json:"status"`
	ChecksumSHA256   string     `json:"checksumSha256,omitempty"`
	UploadedByUserID string     `json:"uploadedByUserId,omitempty"`
	VerifiedAt       *time.Time `json:"verifiedAt,omitempty"`
	DeletedAt        *time.Time `json:"deletedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type InitiateUploadInput struct {
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	Kind         string `json:"kind"`
	FileName     string `json:"fileName"`
	ContentType  string `json:"contentType"`
	SizeBytes    int64  `json:"sizeBytes"`
}

type UploadIntent struct {
	Document  Document          `json:"document"`
	UploadURL string            `json:"uploadUrl"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time         `json:"expiresAt"`
}

type DownloadGrant struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Filter struct {
	ResourceType string
	ResourceID   string
}

type ObjectInfo struct {
	SizeBytes   int64
	ContentType string
	ETag        string
}

type Storage interface {
	PresignPut(ctx context.Context, key, contentType string, expires time.Duration) (string, map[string]string, error)
	Head(ctx context.Context, key string) (ObjectInfo, error)
	PresignGet(ctx context.Context, key string, expires time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

type Repository interface {
	CreatePending(ctx context.Context, document Document) (Document, error)
	Get(ctx context.Context, organizationID, documentID string) (Document, error)
	List(ctx context.Context, organizationID string, filter Filter) ([]Document, error)
	ResourceExists(ctx context.Context, organizationID, resourceType, resourceID string) (bool, error)
	UpdateStatus(ctx context.Context, organizationID, documentID, actorUserID, status string, verifiedAt, deletedAt *time.Time) (Document, error)
}

var (
	ErrNotFound             = errors.New("document not found")
	ErrResourceNotFound     = errors.New("document resource not found")
	ErrInvalidInput         = errors.New("invalid document input")
	ErrInvalidContentType   = errors.New("document content type is not allowed")
	ErrFileTooLarge         = errors.New("document exceeds the 25 MB limit")
	ErrStorageUnavailable   = errors.New("document storage is not configured")
	ErrUploadNotPending     = errors.New("document upload is not pending")
	ErrObjectMismatch       = errors.New("uploaded object does not match declared size or content type")
	ErrDocumentNotAvailable = errors.New("document is not available for download")
)
