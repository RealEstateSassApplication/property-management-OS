package onboarding

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	MaxProperties = 500
	MaxPreviewRows = 30
)

var (
	ErrInvalidHeader = errors.New("invalid property CSV header")
	ErrInvalidCSV = errors.New("invalid property CSV")
	ErrTooManyProperties = errors.New("CSV exceeds 500 property rows")
	ErrNoProperties = errors.New("CSV contains no property rows")
)

type Issue struct {
	Row int `json:"row"`
	Field string `json:"field"`
	Code string `json:"code"`
	Message string `json:"message"`
}

type PropertyRow struct {
	Row int `json:"row"`
	ReferenceCode string `json:"referenceCode,omitempty"`
	Name string `json:"name"`
	PropertyType string `json:"propertyType"`
	AddressLine1 string `json:"addressLine1"`
	City string `json:"city"`
	Region string `json:"region,omitempty"`
	CountryCode string `json:"countryCode"`
	ExternalAvaraPropertyID string `json:"externalAvaraPropertyId,omitempty"`
}

type Report struct {
	TotalRows int `json:"totalRows"`
	ReadyRows int `json:"readyRows"`
	InvalidRows int `json:"invalidRows"`
	PreviewRows []PropertyRow `json:"previewRows"`
	Issues []Issue `json:"issues"`
	CanImport bool `json:"canImport"`
}

// PreviewPropertiesCSV validates a spreadsheet-exported CSV without writing data.
// Row numbers are logical CSV record numbers including the header, not physical
// file lines (quoted fields may span lines).
func PreviewPropertiesCSV(input io.Reader) (Report, error) {
	result := Report{PreviewRows: []PropertyRow{}, Issues: []Issue{}}
	reader := csv.NewReader(input)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return result, fmt.Errorf("%w: could not read header", ErrInvalidHeader)
	}
	index := make(map[string]int, len(header))
	allowed := map[string]bool{
		"referencecode": true, "name": true, "propertytype": true,
		"addressline1": true, "city": true, "region": true,
		"countrycode": true, "externalavarapropertyid": true,
	}
	for column, label := range header {
		key := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(label, "\ufeff")))
		if !allowed[key] || key == "" {
			return result, fmt.Errorf("%w: unexpected column %q", ErrInvalidHeader, label)
		}
		if _, duplicate := index[key]; duplicate {
			return result, fmt.Errorf("%w: duplicate column %q", ErrInvalidHeader, label)
		}
		index[key] = column
	}
	for _, key := range []string{"name", "propertytype", "addressline1", "city", "countrycode"} {
		if _, exists := index[key]; !exists {
			return result, fmt.Errorf("%w: missing required column %s", ErrInvalidHeader, key)
		}
	}

	get := func(record []string, key string) string {
		if column, ok := index[key]; ok {
			return strings.TrimSpace(record[column])
		}
		return ""
	}
	references := make(map[string]int)
	avaraIDs := make(map[string]int)
	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) { break }
		if readErr != nil { return Report{}, fmt.Errorf("%w: malformed row: %v", ErrInvalidCSV, readErr) }
		result.TotalRows++
		if result.TotalRows > MaxProperties { return Report{}, ErrTooManyProperties }
		row := result.TotalRows + 1
		if len(record) != len(header) {
			result.Issues = append(result.Issues, Issue{row, "", "column_count", "Column count differs from CSV header"})
			result.InvalidRows++
			continue
		}
		entry := PropertyRow{
			Row: row,
			ReferenceCode: get(record, "referencecode"),
			Name: get(record, "name"),
			PropertyType: strings.ToLower(get(record, "propertytype")),
			AddressLine1: get(record, "addressline1"),
			City: get(record, "city"),
			Region: get(record, "region"),
			CountryCode: strings.ToUpper(get(record, "countrycode")),
			ExternalAvaraPropertyID: get(record, "externalavarapropertyid"),
		}
		issuesBefore := len(result.Issues)
		for _, item := range []struct{ field, value string }{
			{"name", entry.Name}, {"addressLine1", entry.AddressLine1}, {"city", entry.City},
			{"propertyType", entry.PropertyType}, {"countryCode", entry.CountryCode},
		} {
			if item.value == "" {
				result.Issues = append(result.Issues, Issue{row, item.field, "required", item.field + " is required"})
			}
		}
		for _, item := range []struct{ field, value string }{
			{"referenceCode", entry.ReferenceCode}, {"name", entry.Name},
			{"propertyType", entry.PropertyType}, {"addressLine1", entry.AddressLine1},
			{"city", entry.City}, {"region", entry.Region}, {"countryCode", entry.CountryCode},
			{"externalAvaraPropertyId", entry.ExternalAvaraPropertyID},
		} {
			if !utf8.ValidString(item.value) || strings.ContainsRune(item.value, 0) {
				result.Issues = append(result.Issues, Issue{row, item.field, "invalid_text", "Field contains invalid text"})
			} else if utf8.RuneCountInString(item.value) > 200 {
				result.Issues = append(result.Issues, Issue{row, item.field, "too_long", "Field exceeds 200 characters"})
			}
		}
		switch entry.PropertyType {
		case "", "house", "apartment", "building", "commercial", "land", "other":
		default:
			result.Issues = append(result.Issues, Issue{row, "propertyType", "invalid_type", "Use house, apartment, building, commercial, land or other"})
		}
		if entry.CountryCode != "" {
			valid := len(entry.CountryCode) == 2
			for _, c := range entry.CountryCode {
				if c < 'A' || c > 'Z' { valid = false }
			}
			if !valid { result.Issues = append(result.Issues, Issue{row, "countryCode", "invalid_country", "Use a two-letter ISO country code, for example LK"}) }
		}
		if entry.ReferenceCode != "" {
			key := strings.ToLower(entry.ReferenceCode)
			if first, exists := references[key]; exists {
				result.Issues = append(result.Issues, Issue{row, "referenceCode", "duplicate_reference", fmt.Sprintf("Reference code also occurs on row %d", first)})
			} else { references[key] = row }
		}
		if entry.ExternalAvaraPropertyID != "" {
			key := strings.ToLower(entry.ExternalAvaraPropertyID)
			if first, exists := avaraIDs[key]; exists {
				result.Issues = append(result.Issues, Issue{row, "externalAvaraPropertyId", "duplicate_avara_id", fmt.Sprintf("Avara property ID also occurs on row %d", first)})
			} else { avaraIDs[key] = row }
		}
		if len(result.Issues) == issuesBefore {
			result.ReadyRows++
			if len(result.PreviewRows) < MaxPreviewRows { result.PreviewRows = append(result.PreviewRows, entry) }
		} else {
			result.InvalidRows++
		}
	}
	if result.TotalRows == 0 { return Report{}, ErrNoProperties }
	result.CanImport = result.InvalidRows == 0
	return result, nil
}
