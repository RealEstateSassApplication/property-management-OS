package documents

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	document Document
	exists   bool
}

func (f *fakeRepository) CreatePending(_ context.Context, document Document) (Document, error) { f.document = document; return document, nil }
func (f *fakeRepository) Get(_ context.Context, _, _ string) (Document, error) { if f.document.ID == "" { return Document{}, ErrNotFound }; return f.document, nil }
func (f *fakeRepository) List(context.Context, string, Filter) ([]Document, error) { return []Document{f.document}, nil }
func (f *fakeRepository) ResourceExists(context.Context, string, string, string) (bool, error) { return f.exists, nil }
func (f *fakeRepository) UpdateStatus(_ context.Context, _, _, _ , status string, verifiedAt, deletedAt *time.Time) (Document, error) { f.document.Status = status; f.document.VerifiedAt = verifiedAt; f.document.DeletedAt = deletedAt; return f.document, nil }

type fakeStorage struct {
	info ObjectInfo
	err  error
}

func (f fakeStorage) PresignPut(context.Context, string, string, time.Duration) (string, map[string]string, error) { return "https://upload.example", map[string]string{"Content-Type":"application/pdf"}, f.err }
func (f fakeStorage) Head(context.Context, string) (ObjectInfo, error) { return f.info, f.err }
func (f fakeStorage) PresignGet(context.Context, string, time.Duration) (string, error) { return "https://download.example", f.err }
func (f fakeStorage) Delete(context.Context, string) error { return f.err }

func TestInitiateUploadNormalizesFileAndCreatesPendingDocument(t *testing.T) {
	repository := &fakeRepository{exists: true}
	service := NewService(repository, fakeStorage{})
	intent, err := service.InitiateUpload(context.Background(), "org-1", "user-1", InitiateUploadInput{ResourceType:"lease", ResourceID:"lease-1", Kind:"lease", FileName:"../Signed Lease 2026.pdf", ContentType:"application/pdf; charset=binary", SizeBytes:1024})
	if err != nil { t.Fatal(err) }
	if intent.Document.FileName != "Signed_Lease_2026.pdf" { t.Fatalf("unexpected filename %q", intent.Document.FileName) }
	if intent.Document.Status != "pending" || intent.UploadURL == "" { t.Fatalf("expected pending upload intent: %+v", intent) }
}

func TestCompleteUploadQuarantinesMetadataMismatch(t *testing.T) {
	repository := &fakeRepository{document: Document{ID:"doc-1", OrganizationID:"org-1", Status:"pending", StorageKey:"key", SizeBytes:100, ContentType:"application/pdf"}}
	service := NewService(repository, fakeStorage{info:ObjectInfo{SizeBytes:99, ContentType:"application/pdf"}})
	_, err := service.CompleteUpload(context.Background(), "org-1", "user-1", "doc-1")
	if !errors.Is(err, ErrObjectMismatch) { t.Fatalf("expected mismatch, got %v", err) }
	if repository.document.Status != "quarantined" { t.Fatalf("expected quarantined, got %s", repository.document.Status) }
}

func TestDownloadRequiresAvailableDocument(t *testing.T) {
	repository := &fakeRepository{document: Document{ID:"doc-1", OrganizationID:"org-1", Status:"pending", StorageKey:"key"}}
	service := NewService(repository, fakeStorage{})
	_, err := service.Download(context.Background(), "org-1", "doc-1")
	if !errors.Is(err, ErrDocumentNotAvailable) { t.Fatalf("expected unavailable, got %v", err) }
}
