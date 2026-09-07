package maintenance

import (
	"context"
	"testing"
	"time"
)

type fakeRepository struct {
	vendorInput   CreateVendorInput
	requestInput  CreateRequestInput
	quoteInput    CreateQuoteInput
	evidenceInput CreateEvidenceInput
	requestActor  string
	evidenceActor string
}

func (f *fakeRepository) ListVendors(context.Context, string) ([]Vendor, error) { return nil, nil }
func (f *fakeRepository) CreateVendor(_ context.Context, _ string, input CreateVendorInput) (Vendor, error) {
	f.vendorInput = input
	return Vendor{Name: input.Name, Trade: input.Trade, Email: input.Email, Status: input.Status}, nil
}
func (f *fakeRepository) ListRequests(context.Context, string) ([]Request, error) { return nil, nil }
func (f *fakeRepository) CreateRequest(_ context.Context, _ string, actor string, input CreateRequestInput) (Request, error) {
	f.requestActor = actor
	f.requestInput = input
	return Request{Title: input.Title, Priority: input.Priority}, nil
}
func (f *fakeRepository) UpdateRequestStatus(context.Context, string, string, string) (Request, error) {
	return Request{}, nil
}
func (f *fakeRepository) ListWorkOrders(context.Context, string) ([]WorkOrder, error) {
	return nil, nil
}
func (f *fakeRepository) CreateWorkOrder(context.Context, string, CreateWorkOrderInput, *time.Time) (WorkOrder, error) {
	return WorkOrder{}, nil
}
func (f *fakeRepository) UpdateWorkOrderStatus(context.Context, string, string, string) (WorkOrder, error) {
	return WorkOrder{}, nil
}
func (f *fakeRepository) ListQuotes(context.Context, string) ([]Quote, error) { return nil, nil }
func (f *fakeRepository) CreateQuote(_ context.Context, _ string, input CreateQuoteInput) (Quote, error) {
	f.quoteInput = input
	return Quote{AmountMinor: input.AmountMinor, Currency: input.Currency}, nil
}
func (f *fakeRepository) DecideQuote(context.Context, string, string, string, string) (Quote, error) {
	return Quote{}, nil
}
func (f *fakeRepository) ListEvidence(context.Context, string) ([]Evidence, error) { return nil, nil }
func (f *fakeRepository) CreateEvidence(_ context.Context, _ string, actor string, input CreateEvidenceInput) (Evidence, error) {
	f.evidenceActor = actor
	f.evidenceInput = input
	return Evidence{WorkOrderID: input.WorkOrderID, EvidenceType: input.EvidenceType, Note: input.Note}, nil
}

func TestCreateVendorNormalizesFields(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.CreateVendor(context.Background(), "org", CreateVendorInput{
		Name: "  Colombo Fixers  ", Trade: " PLUMBING ", Email: " TEAM@EXAMPLE.COM ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.vendorInput.Name != "Colombo Fixers" || repo.vendorInput.Trade != "plumbing" || repo.vendorInput.Email != "team@example.com" || repo.vendorInput.Status != "active" {
		t.Fatalf("unexpected normalized vendor: %#v", repo.vendorInput)
	}
}

func TestCreateRequestDefaultsPriorityAndPreservesActor(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.CreateRequest(context.Background(), "org", " user-1 ", CreateRequestInput{
		PropertyID: "property", Title: " Leak ", Description: " Kitchen tap leak ", Category: " PLUMBING ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.requestInput.Priority != "normal" || repo.requestInput.Category != "plumbing" || repo.requestActor != "user-1" {
		t.Fatalf("unexpected request normalization: %#v actor=%q", repo.requestInput, repo.requestActor)
	}
}

func TestCreateQuoteRejectsNonPositiveAmount(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.CreateQuote(context.Background(), "org", CreateQuoteInput{
		WorkOrderID: "wo", VendorID: "vendor", AmountMinor: 0, Currency: "LKR", ScopeSummary: "Repair",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.quoteInput.WorkOrderID != "" {
		t.Fatal("repository should not be called")
	}
}

func TestCreateEvidenceRequiresContent(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.CreateEvidence(context.Background(), "org", "user", CreateEvidenceInput{
		WorkOrderID: "wo", EvidenceType: "note",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.evidenceInput.WorkOrderID != "" {
		t.Fatal("repository should not be called")
	}
}
