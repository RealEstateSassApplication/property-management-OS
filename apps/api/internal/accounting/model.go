package accounting

import (
	"errors"
	"time"
)

type RentAdjustment struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organizationId"`
	ObligationID    string    `json:"obligationId"`
	LeaseReference  string    `json:"leaseReference"`
	TenantName      string    `json:"tenantName"`
	AdjustmentType  string    `json:"adjustmentType"`
	AmountMinor     int64     `json:"amountMinor"`
	Currency        string    `json:"currency"`
	Reason          string    `json:"reason"`
	CreatedByUserID string    `json:"createdByUserId,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

type PaymentReversal struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organizationId"`
	PaymentID        string    `json:"paymentId"`
	TenantName       string    `json:"tenantName"`
	AmountMinor      int64     `json:"amountMinor"`
	Currency         string    `json:"currency"`
	ReferenceCode    string    `json:"referenceCode,omitempty"`
	Reason           string    `json:"reason"`
	ReversedByUserID string    `json:"reversedByUserId,omitempty"`
	ReversedAt       time.Time `json:"reversedAt"`
}

type DepositAccount struct {
	ID                  string    `json:"id"`
	OrganizationID      string    `json:"organizationId"`
	LeaseID             string    `json:"leaseId"`
	LeaseReference      string    `json:"leaseReference"`
	TenantName          string    `json:"tenantName"`
	PropertyName        string    `json:"propertyName"`
	UnitLabel           string    `json:"unitLabel"`
	RequiredAmountMinor int64     `json:"requiredAmountMinor"`
	HeldAmountMinor     int64     `json:"heldAmountMinor"`
	Currency            string    `json:"currency"`
	CreatedAt           time.Time `json:"createdAt"`
}

type DepositTransaction struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organizationId"`
	DepositAccountID string    `json:"depositAccountId"`
	TransactionType  string    `json:"transactionType"`
	AmountMinor      int64     `json:"amountMinor"`
	OccurredOn       string    `json:"occurredOn"`
	Note             string    `json:"note"`
	CreatedByUserID  string    `json:"createdByUserId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
}

type PropertyExpense struct {
	ID              string     `json:"id"`
	OrganizationID  string     `json:"organizationId"`
	PropertyID      string     `json:"propertyId"`
	PropertyName    string     `json:"propertyName"`
	VendorID        string     `json:"vendorId,omitempty"`
	VendorName      string     `json:"vendorName,omitempty"`
	WorkOrderID     string     `json:"workOrderId,omitempty"`
	Category        string     `json:"category"`
	AmountMinor     int64      `json:"amountMinor"`
	Currency        string     `json:"currency"`
	IncurredOn      string     `json:"incurredOn"`
	Note            string     `json:"note"`
	ReferenceCode   string     `json:"referenceCode,omitempty"`
	CreatedByUserID string     `json:"createdByUserId,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	Reversed        bool       `json:"reversed"`
	ReversalReason  string     `json:"reversalReason,omitempty"`
	ReversedAt      *time.Time `json:"reversedAt,omitempty"`
}

type StatementLine struct {
	Date         string `json:"date"`
	LineType     string `json:"lineType"`
	PropertyID   string `json:"propertyId"`
	PropertyName string `json:"propertyName"`
	Description  string `json:"description"`
	Currency     string `json:"currency"`
	GrossMinor   int64  `json:"grossMinor"`
	OwnershipBPS int    `json:"ownershipBps"`
	OwnerMinor   int64  `json:"ownerMinor"`
}

type StatementCurrencySummary struct {
	Currency            string `json:"currency"`
	IncomeMinor         int64  `json:"incomeMinor"`
	ExpenseMinor        int64  `json:"expenseMinor"`
	NetOwnerAmountMinor int64  `json:"netOwnerAmountMinor"`
}

type OwnerStatement struct {
	OwnerID   string                     `json:"ownerId"`
	OwnerName string                     `json:"ownerName"`
	From      string                     `json:"from"`
	To        string                     `json:"to"`
	Summaries []StatementCurrencySummary `json:"summaries"`
	Lines     []StatementLine            `json:"lines"`
}

type CreateRentAdjustmentInput struct {
	ObligationID   string `json:"obligationId"`
	AdjustmentType string `json:"adjustmentType"`
	AmountMinor    int64  `json:"amountMinor"`
	Reason         string `json:"reason"`
}

type ReversePaymentInput struct {
	PaymentID string `json:"paymentId"`
	Reason    string `json:"reason"`
}

type CreateDepositAccountInput struct {
	LeaseID string `json:"leaseId"`
}

type CreateDepositTransactionInput struct {
	DepositAccountID string `json:"depositAccountId"`
	TransactionType  string `json:"transactionType"`
	AmountMinor      int64  `json:"amountMinor"`
	OccurredOn       string `json:"occurredOn"`
	Note             string `json:"note"`
}

type CreateExpenseInput struct {
	PropertyID    string `json:"propertyId"`
	VendorID      string `json:"vendorId"`
	WorkOrderID   string `json:"workOrderId"`
	Category      string `json:"category"`
	AmountMinor   int64  `json:"amountMinor"`
	Currency      string `json:"currency"`
	IncurredOn    string `json:"incurredOn"`
	Note          string `json:"note"`
	ReferenceCode string `json:"referenceCode"`
}

type ReverseExpenseInput struct {
	ExpenseID string `json:"expenseId"`
	Reason    string `json:"reason"`
}

var (
	ErrObligationNotFound       = errors.New("rent obligation not found")
	ErrAdjustmentWouldOverpay   = errors.New("credit or write-off exceeds outstanding balance")
	ErrPaymentNotFound          = errors.New("payment not found")
	ErrPaymentAlreadyReversed   = errors.New("payment already reversed")
	ErrPaymentNotPosted         = errors.New("only posted payments can be reversed")
	ErrLeaseNotFound            = errors.New("lease not found")
	ErrDepositAccountExists     = errors.New("security deposit account already exists for lease")
	ErrDepositAccountNotFound   = errors.New("security deposit account not found")
	ErrDepositInsufficientFunds = errors.New("security deposit transaction exceeds held funds")
	ErrPropertyNotFound         = errors.New("property not found")
	ErrVendorNotFound           = errors.New("vendor not found")
	ErrWorkOrderNotFound        = errors.New("work order not found")
	ErrExpenseNotFound          = errors.New("property expense not found")
	ErrExpenseAlreadyReversed   = errors.New("property expense already reversed")
	ErrOwnerNotFound            = errors.New("owner not found")
)
