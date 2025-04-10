package converters

import "github.com/TerraDharitri/drt-go-chain-es-indexer/data"

// TruncateFieldIfExceedsMaxLength will truncate the provided field if the max length exceeds
func TruncateFieldIfExceedsMaxLength(field string) string {
	if len(field) > data.MaxFieldLength {
		return field[:data.MaxFieldLength]
	}

	return field
}

// TruncateFieldIfExceedsMaxLengthBase64 will truncate the provided field if the max length after base64 encoding would be exceeded
// this function will be used for the fields that after will be base64 encoded
func TruncateFieldIfExceedsMaxLengthBase64(field string) string {
	if len(field) > data.MaxKeywordFieldLengthBeforeBase64Encoding {
		return field[:data.MaxKeywordFieldLengthBeforeBase64Encoding]
	}

	return field
}

// TruncateSliceElementsIfExceedsMaxLength will truncate the provided slice of the field if the max length is exceeded
func TruncateSliceElementsIfExceedsMaxLength(fields []string) []string {
	var localFields []string
	for _, field := range fields {
		localFields = append(localFields, TruncateFieldIfExceedsMaxLength(field))
	}

	return localFields
}
