package accounting

import (
	"context"
	"testing"
	"time"
)

type fakeRepository struct{}

func (fakeRepository) ListRentAdjustments(context.Context, string) ([]RentAdjustment, error) {
	return nil, nil
}
func (fakeRepository) CreateRentAdjustment(_ context.Context, _ string, _ string, input CreateRentAdjustmentInput) (RentAdjustment, error) {
	return RentAdjustment{ObligationID: input.ObligationID, AdjustmentType: input.AdjustmentType, AmountMinor: input.AmountMinor, Reason: input.Reason}, nil
}
func (fakeRepository) ListPaymentReversals(context.Context, string) ([]PaymentReversal, error) {
	return nil, nil
}
func (fakeRepository) ReversePayment(_ context.Context, _ string, _ string, input ReversePaymentInput) (PaymentReversal, error) {
	return PaymentReversal{PaymentID: input.PaymentID, Reason: input.Reason}, nil
}
func (fakeRepository) ListDepositAccounts(context.Context, string) ([]DepositAccount, error) {
	return nil, nil
}
func (fakeRepository) CreateDepositAccount(_ context.Context, _ string, input CreateDepositAccountInput) (DepositAccount, error) {
	return DepositAccount{LeaseID: input.LeaseID}, nil
}
func (fakeRepository) ListDepositTransactions(context.Context, string) ([]DepositTransaction, error) {
	return nil, nil
}
func (fakeRepository) CreateDepositTransaction(_ context.Context, _ string, _ string, input CreateDepositTransactionInput) (DepositTransaction, error) {
	return DepositTransaction{DepositAccountID: input.DepositAccountID, TransactionType: input.TransactionType, AmountMinor: input.AmountMinor, OccurredOn: input.OccurredOn, Note: input.Note}, nil
}
func (fakeRepository) ListExpenses(context.Context, string) ([]PropertyExpense, error) {
	return nil, nil
}
func (fakeRepository) CreateExpense(_ context.Context, _ string, _ string, input CreateExpenseInput) (PropertyExpense, error) {
	return PropertyExpense{PropertyID: input.PropertyID, Category: input.Category, AmountMinor: input.AmountMinor, Currency: input.Currency, IncurredOn: input.IncurredOn, Note: input.Note}, nil
}
func (fakeRepository) ReverseExpense(_ context.Context, _ string, _ string, input ReverseExpenseInput) (PropertyExpense, error) {
	return PropertyExpense{ID: input.ExpenseID, ReversalReason: input.Reason, Reversed: true}, nil
}
func (fakeRepository) OwnerStatement(_ context.Context, _ string, ownerID string, from, to time.Time) (OwnerStatement, error) {
	return OwnerStatement{OwnerID: ownerID, From: from.Format("2006-01-02"), To: to.Format("2006-01-02")}, nil
}

func TestCreateRentAdjustmentNormalizesInput(t *testing.T) {
	service := NewService(fakeRepository{})
	item, err := service.CreateRentAdjustment(context.Background(), "org", "actor", CreateRentAdjustmentInput{ObligationID: " obligation ", AdjustmentType: " LATE_FEE ", AmountMinor: 150000, Reason: " Late payment "})
	if err != nil {
		t.Fatal(err)
	}
	if item.ObligationID != "obligation" || item.AdjustmentType != "late_fee" || item.Reason != "Late payment" {
		t.Fatalf("unexpected normalized adjustment: %+v", item)
	}
}

func TestCreateDepositTransactionRejectsInvalidDate(t *testing.T) {
	service := NewService(fakeRepository{})
	_, err := service.CreateDepositTransaction(context.Background(), "org", "actor", CreateDepositTransactionInput{DepositAccountID: "deposit", TransactionType: "received", AmountMinor: 1000, OccurredOn: "07/09/2026", Note: "Received"})
	if err == nil {
		t.Fatal("expected invalid date to fail")
	}
}

func TestCreateExpenseNormalizesCurrency(t *testing.T) {
	service := NewService(fakeRepository{})
	item, err := service.CreateExpense(context.Background(), "org", "actor", CreateExpenseInput{PropertyID: "property", Category: " Maintenance ", AmountMinor: 250000, Currency: "lkr", IncurredOn: "2026-09-07", Note: " Repair invoice "})
	if err != nil {
		t.Fatal(err)
	}
	if item.Category != "maintenance" || item.Currency != "LKR" || item.Note != "Repair invoice" {
		t.Fatalf("unexpected normalized expense: %+v", item)
	}
}

func TestOwnerStatementRejectsReverseRange(t *testing.T) {
	service := NewService(fakeRepository{})
	_, err := service.OwnerStatement(context.Background(), "org", "owner", "2026-09-30", "2026-09-01")
	if err == nil {
		t.Fatal("expected reversed statement range to fail")
	}
}
