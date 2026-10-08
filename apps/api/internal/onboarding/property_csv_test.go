package onboarding

import (
	"errors"
	"strings"
	"testing"
)

const header = "referenceCode,name,propertyType,addressLine1,city,region,countryCode,externalAvaraPropertyId\n"

func TestPreviewAcceptsQuotedExcelCSVAndBOM(t *testing.T) {
	csv := "\ufeff" + header + "P-01,\"Villa, Ocean\",HOUSE,\"12 Main Road, South\",Galle,South,lk,avara-1\n"
	got, err := PreviewPropertiesCSV(strings.NewReader(csv))
	if err != nil { t.Fatal(err) }
	if !got.CanImport || got.TotalRows != 1 || got.ReadyRows != 1 || got.InvalidRows != 0 { t.Fatalf("unexpected report: %+v", got) }
	p := got.PreviewRows[0]
	if p.Name != "Villa, Ocean" || p.CountryCode != "LK" || p.PropertyType != "house" || p.ReferenceCode != "P-01" { t.Fatalf("normalization failed: %+v", p) }
}

func TestPreviewRejectsDuplicateReferenceAndAvaraID(t *testing.T) {
	data := header + "P-01,Home,house,Address,Colombo,West,LK,avara-1\n" +
		"p-01,House,house,Other Road,Kandy,Central,LK,AVARA-1\n"
	got, err := PreviewPropertiesCSV(strings.NewReader(data))
	if err != nil { t.Fatal(err) }
	if got.CanImport || got.ReadyRows != 1 || got.InvalidRows != 1 || len(got.Issues) != 2 { t.Fatalf("duplicate not caught: %+v", got) }
	for _, issue := range got.Issues { if issue.Row != 3 { t.Fatalf("unexpected logical row %d", issue.Row) } }
}

func TestPreviewRejectsInvalidRowWithoutStoppingValidRows(t *testing.T) {
	data := header + "P-01,Home,house,Address,Colombo,West,LK,a-1\n" +
		"P-02,,castle,,Kandy,Central,123,a-2\n"
	got, err := PreviewPropertiesCSV(strings.NewReader(data))
	if err != nil { t.Fatal(err) }
	if got.CanImport || got.ReadyRows != 1 || got.InvalidRows != 1 || len(got.Issues) < 3 { t.Fatalf("invalid row not reported: %+v", got) }
}

func TestPreviewRejectsMalformedCSVAndMissingHeader(t *testing.T) {
	tests := []struct {input string; target error}{
		{"name,propertyType,addressLine1,city\nOne,house,Road,Colombo\n", ErrInvalidHeader},
		{header + "P-01,\"unclosed,house,Road,City,Region,LK,a-1\n", ErrInvalidCSV},
		{header, ErrNoProperties},
	}
	for _, tc := range tests {
		_, err := PreviewPropertiesCSV(strings.NewReader(tc.input))
		if !errors.Is(err, tc.target) { t.Errorf("expected %v, got %v", tc.target, err) }
	}
}

func TestPreviewRejectsUnevenAndDuplicateColumns(t *testing.T) {
	got, err := PreviewPropertiesCSV(strings.NewReader(header + "P-01,One,house,Address,Colombo\n"))
	if err != nil || got.ReadyRows != 0 || got.InvalidRows != 1 || got.Issues[0].Code != "column_count" { t.Fatalf("unexpected uneven-row result: %+v, %v", got, err) }
	_, err = PreviewPropertiesCSV(strings.NewReader("name,NAME,propertyType,addressLine1,city,countryCode\n"))
	if !errors.Is(err, ErrInvalidHeader) { t.Fatalf("duplicate column accepted: %v", err) }
}

func TestPreviewRejectsMoreThanFiveHundred(t *testing.T) {
	var b strings.Builder
	b.WriteString(header)
	for i := 0; i < 501; i++ { b.WriteString("P-01,Home,house,Address,Colombo,West,LK,\n") }
	_, err := PreviewPropertiesCSV(strings.NewReader(b.String()))
	if !errors.Is(err, ErrTooManyProperties) { t.Fatalf("expected row limit, got %v", err) }
}

func TestPreviewLimitsPreviewRows(t *testing.T) {
	var b strings.Builder
	b.WriteString(header)
	for i := 0; i < 40; i++ { b.WriteString("P-,Home,house,Address,Colombo,West,LK,\n") }
	got, err := PreviewPropertiesCSV(strings.NewReader(b.String()))
	if err != nil { t.Fatal(err) }
	if got.ReadyRows != 40 || len(got.PreviewRows) != MaxPreviewRows { t.Fatalf("preview truncation wrong: %+v", got) }
}

func TestPreviewRejectsUnsafeText(t *testing.T) {
	got, err := PreviewPropertiesCSV(strings.NewReader(header + "P-01,Home\x00,house,Address,Colombo,West,LK,\n"))
	if err != nil { t.Fatal(err) }
	if got.CanImport || got.InvalidRows != 1 { t.Fatalf("NUL accepted: %+v", got) }
}
