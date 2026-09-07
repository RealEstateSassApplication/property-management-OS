package accounting

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	ListRentAdjustments(ctx context.Context, organizationID string) ([]RentAdjustment, error)
	CreateRentAdjustment(ctx context.Context, organizationID, actorUserID string, input CreateRentAdjustmentInput) (RentAdjustment, error)
	ListPaymentReversals(ctx context.Context, organizationID string) ([]PaymentReversal, error)
	ReversePayment(ctx context.Context, organizationID, actorUserID string, input ReversePaymentInput) (PaymentReversal, error)
	ListDepositAccounts(ctx context.Context, organizationID string) ([]DepositAccount, error)
	CreateDepositAccount(ctx context.Context, organizationID string, input CreateDepositAccountInput) (DepositAccount, error)
	ListDepositTransactions(ctx context.Context, organizationID string) ([]DepositTransaction, error)
	CreateDepositTransaction(ctx context.Context, organizationID, actorUserID string, input CreateDepositTransactionInput) (DepositTransaction, error)
	ListExpenses(ctx context.Context, organizationID string) ([]PropertyExpense, error)
	CreateExpense(ctx context.Context, organizationID, actorUserID string, input CreateExpenseInput) (PropertyExpense, error)
	ReverseExpense(ctx context.Context, organizationID, actorUserID string, input ReverseExpenseInput) (PropertyExpense, error)
	OwnerStatement(ctx context.Context, organizationID, ownerID string, from, to time.Time) (OwnerStatement, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) ListRentAdjustments(ctx context.Context, organizationID string) ([]RentAdjustment, error) {
	return s.repository.ListRentAdjustments(ctx, organizationID)
}

func (s *Service) CreateRentAdjustment(ctx context.Context, organizationID, actorUserID string, input CreateRentAdjustmentInput) (RentAdjustment, error) {
	input.ObligationID = strings.TrimSpace(input.ObligationID)
	input.AdjustmentType = strings.ToLower(strings.TrimSpace(input.AdjustmentType))
	input.Reason = strings.TrimSpace(input.Reason)
	if input.ObligationID == "" || input.AmountMinor <= 0 || len(input.Reason) < 4 {
		return RentAdjustment{}, fmt.Errorf("obligationId, positive amountMinor and a reason are required")
	}
	if !contains([]string{"charge", "late_fee", "credit", "writeoff"}, input.AdjustmentType) {
		return RentAdjustment{}, fmt.Errorf("unsupported adjustmentType")
	}
	return s.repository.CreateRentAdjustment(ctx, organizationID, actorUserID, input)
}

func (s *Service) ListPaymentReversals(ctx context.Context, organizationID string) ([]PaymentReversal, error) {
	return s.repository.ListPaymentReversals(ctx, organizationID)
}

func (s *Service) ReversePayment(ctx context.Context, organizationID, actorUserID string, input ReversePaymentInput) (PaymentReversal, error) {
	input.PaymentID = strings.TrimSpace(input.PaymentID)
	input.Reason = strings.TrimSpace(input.Reason)
	if input.PaymentID == "" || len(input.Reason) < 4 {
		return PaymentReversal{}, fmt.Errorf("paymentId and reversal reason are required")
	}
	return s.repository.ReversePayment(ctx, organizationID, actorUserID, input)
}

func (s *Service) ListDepositAccounts(ctx context.Context, organizationID string) ([]DepositAccount, error) {
	return s.repository.ListDepositAccounts(ctx, organizationID)
}

func (s *Service) CreateDepositAccount(ctx context.Context, organizationID string, input CreateDepositAccountInput) (DepositAccount, error) {
	input.LeaseID = strings.TrimSpace(input.LeaseID)
	if input.LeaseID == "" {
		return DepositAccount{}, fmt.Errorf("leaseId is required")
	}
	return s.repository.CreateDepositAccount(ctx, organizationID, input)
}

func (s *Service) ListDepositTransactions(ctx context.Context, organizationID string) ([]DepositTransaction, error) {
	return s.repository.ListDepositTransactions(ctx, organizationID)
}

func (s *Service) CreateDepositTransaction(ctx context.Context, organizationID, actorUserID string, input CreateDepositTransactionInput) (DepositTransaction, error) {
	input.DepositAccountID = strings.TrimSpace(input.DepositAccountID)
	input.TransactionType = strings.ToLower(strings.TrimSpace(input.TransactionType))
	input.OccurredOn = strings.TrimSpace(input.OccurredOn)
	input.Note = strings.TrimSpace(input.Note)
	if input.DepositAccountID == "" || input.AmountMinor <= 0 || len(input.Note) < 3 {
		return DepositTransaction{}, fmt.Errorf("depositAccountId, positive amountMinor and note are required")
	}
	if !contains([]string{"received", "deduction", "refund", "adjustment_increase", "adjustment_decrease"}, input.TransactionType) {
		return DepositTransaction{}, fmt.Errorf("unsupported transactionType")
	}
	if _, err := time.Parse("2006-01-02", input.OccurredOn); err != nil {
		return DepositTransaction{}, fmt.Errorf("occurredOn must use YYYY-MM-DD")
	}
	return s.repository.CreateDepositTransaction(ctx, organizationID, actorUserID, input)
}

func (s *Service) ListExpenses(ctx context.Context, organizationID string) ([]PropertyExpense, error) {
	return s.repository.ListExpenses(ctx, organizationID)
}

func (s *Service) CreateExpense(ctx context.Context, organizationID, actorUserID string, input CreateExpenseInput) (PropertyExpense, error) {
	input.PropertyID = strings.TrimSpace(input.PropertyID)
	input.VendorID = strings.TrimSpace(input.VendorID)
	input.WorkOrderID = strings.TrimSpace(input.WorkOrderID)
	input.Category = strings.ToLower(strings.TrimSpace(input.Category))
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.IncurredOn = strings.TrimSpace(input.IncurredOn)
	input.Note = strings.TrimSpace(input.Note)
	input.ReferenceCode = strings.TrimSpace(input.ReferenceCode)
	if input.PropertyID == "" || input.AmountMinor <= 0 || len(input.Note) < 3 {
		return PropertyExpense{}, fmt.Errorf("propertyId, positive amountMinor and note are required")
	}
	if !contains([]string{"maintenance", "utility", "tax", "insurance", "management", "cleaning", "security", "other"}, input.Category) {
		return PropertyExpense{}, fmt.Errorf("unsupported category")
	}
	if len(input.Currency) != 3 {
		return PropertyExpense{}, fmt.Errorf("currency must be a three-letter code")
	}
	if _, err := time.Parse("2006-01-02", input.IncurredOn); err != nil {
		return PropertyExpense{}, fmt.Errorf("incurredOn must use YYYY-MM-DD")
	}
	return s.repository.CreateExpense(ctx, organizationID, actorUserID, input)
}

func (s *Service) ReverseExpense(ctx context.Context, organizationID, actorUserID string, input ReverseExpenseInput) (PropertyExpense, error) {
	input.ExpenseID = strings.TrimSpace(input.ExpenseID)
	input.Reason = strings.TrimSpace(input.Reason)
	if input.ExpenseID == "" || len(input.Reason) < 4 {
		return PropertyExpense{}, fmt.Errorf("expenseId and reversal reason are required")
	}
	return s.repository.ReverseExpense(ctx, organizationID, actorUserID, input)
}

func (s *Service) OwnerStatement(ctx context.Context, organizationID, ownerID, fromText, toText string) (OwnerStatement, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return OwnerStatement{}, fmt.Errorf("ownerId is required")
	}
	from, err := time.Parse("2006-01-02", strings.TrimSpace(fromText))
	if err != nil {
		return OwnerStatement{}, fmt.Errorf("from must use YYYY-MM-DD")
	}
	to, err := time.Parse("2006-01-02", strings.TrimSpace(toText))
	if err != nil {
		return OwnerStatement{}, fmt.Errorf("to must use YYYY-MM-DD")
	}
	if to.Before(from) {
		return OwnerStatement{}, fmt.Errorf("to must be on or after from")
	}
	return s.repository.OwnerStatement(ctx, organizationID, ownerID, from, to)
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
