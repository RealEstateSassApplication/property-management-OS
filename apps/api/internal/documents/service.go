package documents

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"strings"
	"time"
	"unicode"
)

var allowedContentTypes = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"text/plain":      true,
	"text/csv":        true,
}

var allowedKinds = map[string]bool{
	"lease": true, "identity": true, "inspection": true, "invoice": true,
	"receipt": true, "maintenance": true, "statement": true, "other": true,
}

var allowedResourceTypes = map[string]bool{
	"organization": true, "property": true, "unit": true, "tenant": true,
	"lease": true, "owner": true, "rent_payment": true,
	"maintenance_request": true, "work_order": true, "vendor": true,
}

type Service struct {
	repository Repository
	storage    Storage
	now        func() time.Time
}

func NewService(repository Repository, storage Storage) *Service {
	return &Service{repository: repository, storage: storage, now: time.Now}
}

func (s *Service) InitiateUpload(ctx context.Context, organizationID, userID string, input InitiateUploadInput) (UploadIntent, error) {
	if s.storage == nil {
		return UploadIntent{}, ErrStorageUnavailable
	}
	input.ResourceType = strings.TrimSpace(strings.ToLower(input.ResourceType))
	input.ResourceID = strings.TrimSpace(input.ResourceID)
	input.Kind = strings.TrimSpace(strings.ToLower(input.Kind))
	input.FileName = sanitizeFileName(input.FileName)
	input.ContentType = normalizeContentType(input.ContentType)
	if !allowedResourceTypes[input.ResourceType] || !allowedKinds[input.Kind] || input.FileName == "" || input.SizeBytes <= 0 {
		return UploadIntent{}, ErrInvalidInput
	}
	if input.SizeBytes > MaxDocumentSize {
		return UploadIntent{}, ErrFileTooLarge
	}
	if !allowedContentTypes[input.ContentType] {
		return UploadIntent{}, ErrInvalidContentType
	}
	if input.ResourceType == "organization" {
		input.ResourceID = organizationID
	} else if input.ResourceID == "" {
		return UploadIntent{}, ErrInvalidInput
	}
	if input.ResourceType != "organization" {
		exists, err := s.repository.ResourceExists(ctx, organizationID, input.ResourceType, input.ResourceID)
		if err != nil {
			return UploadIntent{}, err
		}
		if !exists {
			return UploadIntent{}, ErrResourceNotFound
		}
	}

	id, err := newUUIDv4()
	if err != nil {
		return UploadIntent{}, err
	}
	now := s.now().UTC()
	storageKey := fmt.Sprintf("%s/documents/%04d/%02d/%s-%s", organizationID, now.Year(), now.Month(), id, input.FileName)
	expiresAt := now.Add(10 * time.Minute)
	uploadURL, headers, err := s.storage.PresignPut(ctx, storageKey, input.ContentType, 10*time.Minute)
	if err != nil {
		return UploadIntent{}, err
	}
	document, err := s.repository.CreatePending(ctx, Document{
		ID: id, OrganizationID: organizationID, ResourceType: input.ResourceType, ResourceID: input.ResourceID,
		Kind: input.Kind, FileName: input.FileName, ContentType: input.ContentType, SizeBytes: input.SizeBytes,
		StorageKey: storageKey, Status: "pending", UploadedByUserID: userID,
	})
	if err != nil {
		_ = s.storage.Delete(ctx, storageKey)
		return UploadIntent{}, err
	}
	return UploadIntent{Document: document, UploadURL: uploadURL, Headers: headers, ExpiresAt: expiresAt}, nil
}

func (s *Service) CompleteUpload(ctx context.Context, organizationID, userID, documentID string) (Document, error) {
	if s.storage == nil {
		return Document{}, ErrStorageUnavailable
	}
	document, err := s.repository.Get(ctx, organizationID, documentID)
	if err != nil {
		return Document{}, err
	}
	if document.Status != "pending" {
		return Document{}, ErrUploadNotPending
	}
	info, err := s.storage.Head(ctx, document.StorageKey)
	if err != nil {
		return Document{}, err
	}
	if info.SizeBytes != document.SizeBytes || normalizeContentType(info.ContentType) != document.ContentType {
		now := s.now().UTC()
		_, _ = s.repository.UpdateStatus(ctx, organizationID, documentID, userID, "quarantined", nil, &now)
		return Document{}, ErrObjectMismatch
	}
	now := s.now().UTC()
	return s.repository.UpdateStatus(ctx, organizationID, documentID, userID, "available", &now, nil)
}

func (s *Service) List(ctx context.Context, organizationID string, filter Filter) ([]Document, error) {
	filter.ResourceType = strings.TrimSpace(strings.ToLower(filter.ResourceType))
	filter.ResourceID = strings.TrimSpace(filter.ResourceID)
	return s.repository.List(ctx, organizationID, filter)
}

func (s *Service) Download(ctx context.Context, organizationID, documentID string) (DownloadGrant, error) {
	if s.storage == nil {
		return DownloadGrant{}, ErrStorageUnavailable
	}
	document, err := s.repository.Get(ctx, organizationID, documentID)
	if err != nil {
		return DownloadGrant{}, err
	}
	if document.Status != "available" {
		return DownloadGrant{}, ErrDocumentNotAvailable
	}
	url, err := s.storage.PresignGet(ctx, document.StorageKey, 5*time.Minute)
	if err != nil {
		return DownloadGrant{}, err
	}
	return DownloadGrant{URL: url, ExpiresAt: s.now().UTC().Add(5 * time.Minute)}, nil
}

func (s *Service) Delete(ctx context.Context, organizationID, userID, documentID string) (Document, error) {
	document, err := s.repository.Get(ctx, organizationID, documentID)
	if err != nil {
		return Document{}, err
	}
	if document.Status == "deleted" {
		return document, nil
	}
	if s.storage != nil {
		if err := s.storage.Delete(ctx, document.StorageKey); err != nil {
			return Document{}, err
		}
	}
	now := s.now().UTC()
	return s.repository.UpdateStatus(ctx, organizationID, documentID, userID, "deleted", nil, &now)
}

func normalizeContentType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if i := strings.Index(value, ";"); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}
	return value
}

func sanitizeFileName(value string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	value = path.Base(strings.TrimSpace(value))
	if value == "." || value == "/" || value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) {
			b.WriteByte('_')
		}
		if b.Len() >= 180 {
			break
		}
	}
	return strings.Trim(b.String(), ".")
}

func newUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32]), nil
}
