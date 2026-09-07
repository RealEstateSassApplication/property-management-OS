package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/accounting"
)

type accountingHandler struct{ service *accounting.Service }

func (h accountingHandler) listAdjustments(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListRentAdjustments(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "accounting_adjustments_list_failed", "could not load rent adjustments")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h accountingHandler) createAdjustment(w http.ResponseWriter, r *http.Request) {
	var input accounting.CreateRentAdjustmentInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	actorUserID, _ := userIDFromContext(r.Context())
	item, err := h.service.CreateRentAdjustment(r.Context(), organizationID, actorUserID, input)
	switch {
	case errors.Is(err, accounting.ErrObligationNotFound):
		writeError(w, http.StatusNotFound, "rent_obligation_not_found", err.Error())
		return
	case errors.Is(err, accounting.ErrAdjustmentWouldOverpay):
		writeError(w, http.StatusConflict, "rent_adjustment_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "rent_adjustment_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h accountingHandler) listReversals(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListPaymentReversals(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "payment_reversals_list_failed", "could not load payment reversals")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h accountingHandler) reversePayment(w http.ResponseWriter, r *http.Request) {
	input := accounting.ReversePaymentInput{PaymentID: r.PathValue("paymentID")}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input.Reason = body.Reason
	organizationID, _ := organizationIDFromContext(r.Context())
	actorUserID, _ := userIDFromContext(r.Context())
	item, err := h.service.ReversePayment(r.Context(), organizationID, actorUserID, input)
	switch {
	case errors.Is(err, accounting.ErrPaymentNotFound):
		writeError(w, http.StatusNotFound, "payment_not_found", err.Error())
		return
	case errors.Is(err, accounting.ErrPaymentAlreadyReversed):
		writeError(w, http.StatusConflict, "payment_already_reversed", err.Error())
		return
	case errors.Is(err, accounting.ErrPaymentNotPosted):
		writeError(w, http.StatusBadRequest, "payment_not_posted", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "payment_reversal_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h accountingHandler) listDepositAccounts(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListDepositAccounts(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "deposit_accounts_list_failed", "could not load security deposits")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h accountingHandler) createDepositAccount(w http.ResponseWriter, r *http.Request) {
	var input accounting.CreateDepositAccountInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateDepositAccount(r.Context(), organizationID, input)
	switch {
	case errors.Is(err, accounting.ErrLeaseNotFound):
		writeError(w, http.StatusNotFound, "lease_not_found", err.Error())
		return
	case errors.Is(err, accounting.ErrDepositAccountExists):
		writeError(w, http.StatusConflict, "deposit_account_exists", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "deposit_account_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h accountingHandler) listDepositTransactions(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListDepositTransactions(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "deposit_transactions_list_failed", "could not load deposit transactions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h accountingHandler) createDepositTransaction(w http.ResponseWriter, r *http.Request) {
	var input accounting.CreateDepositTransactionInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	actorUserID, _ := userIDFromContext(r.Context())
	item, err := h.service.CreateDepositTransaction(r.Context(), organizationID, actorUserID, input)
	switch {
	case errors.Is(err, accounting.ErrDepositAccountNotFound):
		writeError(w, http.StatusNotFound, "deposit_account_not_found", err.Error())
		return
	case errors.Is(err, accounting.ErrDepositInsufficientFunds):
		writeError(w, http.StatusConflict, "deposit_insufficient_funds", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "deposit_transaction_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h accountingHandler) listExpenses(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListExpenses(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "expenses_list_failed", "could not load property expenses")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h accountingHandler) createExpense(w http.ResponseWriter, r *http.Request) {
	var input accounting.CreateExpenseInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	actorUserID, _ := userIDFromContext(r.Context())
	item, err := h.service.CreateExpense(r.Context(), organizationID, actorUserID, input)
	switch {
	case errors.Is(err, accounting.ErrPropertyNotFound), errors.Is(err, accounting.ErrVendorNotFound), errors.Is(err, accounting.ErrWorkOrderNotFound):
		writeError(w, http.StatusNotFound, "expense_resource_not_found", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "expense_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h accountingHandler) reverseExpense(w http.ResponseWriter, r *http.Request) {
	input := accounting.ReverseExpenseInput{ExpenseID: r.PathValue("expenseID")}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input.Reason = body.Reason
	organizationID, _ := organizationIDFromContext(r.Context())
	actorUserID, _ := userIDFromContext(r.Context())
	item, err := h.service.ReverseExpense(r.Context(), organizationID, actorUserID, input)
	switch {
	case errors.Is(err, accounting.ErrExpenseNotFound):
		writeError(w, http.StatusNotFound, "expense_not_found", err.Error())
		return
	case errors.Is(err, accounting.ErrExpenseAlreadyReversed):
		writeError(w, http.StatusConflict, "expense_already_reversed", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "expense_reversal_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h accountingHandler) ownerStatement(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.OwnerStatement(r.Context(), organizationID, r.PathValue("ownerID"), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if errors.Is(err, accounting.ErrOwnerNotFound) {
		writeError(w, http.StatusNotFound, "owner_not_found", err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "owner_statement_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}
