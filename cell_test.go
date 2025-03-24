package sheetah_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nukokusa/sheetah"
)

func AssertDiff(t *testing.T, expected, actual any, opts ...cmp.Option) {
	t.Helper()

	if diff := cmp.Diff(expected, actual, opts...); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestStringCell_Value(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		str      string
		typ      sheetah.ColumnType
		expected any
	}{
		{
			name:     "string",
			str:      "sheetah",
			typ:      sheetah.ColumnTypeString,
			expected: "sheetah",
		},
		{
			name:     "number",
			str:      "12.3",
			typ:      sheetah.ColumnTypeNumber,
			expected: float64(12.3),
		},
		{
			name:     "number: integer",
			str:      "12.0",
			typ:      sheetah.ColumnTypeNumber,
			expected: int64(12),
		},
		{
			name:     "number: invalid",
			str:      "hello",
			typ:      sheetah.ColumnTypeNumber,
			expected: nil,
		},
		{
			name:     "boolean: true",
			str:      "true",
			typ:      sheetah.ColumnTypeBool,
			expected: true,
		},
		{
			name:     "boolean: false",
			str:      "false",
			typ:      sheetah.ColumnTypeBool,
			expected: false,
		},
		{
			name:     "timestamp",
			str:      "2000/01/02 3:04:05",
			typ:      sheetah.ColumnTypeTimestamp,
			expected: time.Date(2000, 1, 2, 3, 4, 5, 0, loc),
		},
		{
			name:     "timestamp: timezone",
			str:      "2000/01/02 3:04:05Z", // UTC
			typ:      sheetah.ColumnTypeTimestamp,
			expected: time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		},
		{
			name:     "timestamp: date",
			str:      "2000/01/02",
			typ:      sheetah.ColumnTypeTimestamp,
			expected: time.Date(2000, 1, 2, 0, 0, 0, 0, loc),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sheetah.NewStringCell(tt.str, loc).Value(tt.typ)
			AssertDiff(t, tt.expected, result)
		})
	}
}

func TestNumberCell_Value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cell     sheetah.NumberCell
		typ      sheetah.ColumnType
		expected any
	}{
		{
			name:     "number",
			cell:     12.3,
			typ:      sheetah.ColumnTypeNumber,
			expected: float64(12.3),
		},
		{
			name:     "number: integer",
			cell:     12.0,
			typ:      sheetah.ColumnTypeNumber,
			expected: int64(12),
		},
		{
			name:     "string",
			cell:     12.3,
			typ:      sheetah.ColumnTypeString,
			expected: "12.3",
		},
		{
			name:     "string: integer",
			cell:     12.0,
			typ:      sheetah.ColumnTypeString,
			expected: "12",
		},
		{
			name:     "boolean: zero",
			cell:     0.0,
			typ:      sheetah.ColumnTypeBool,
			expected: false,
		},
		{
			name:     "boolean: non-zero",
			cell:     1.0,
			typ:      sheetah.ColumnTypeBool,
			expected: true,
		},
		{
			name:     "invalid type",
			cell:     12.3,
			typ:      sheetah.ColumnTypeTimestamp,
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cell.Value(tt.typ)
			AssertDiff(t, tt.expected, result)
		})
	}
}

func TestBoolCell_Value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cell     sheetah.BoolCell
		typ      sheetah.ColumnType
		expected any
	}{
		{
			name:     "boolean: true",
			cell:     true,
			typ:      sheetah.ColumnTypeBool,
			expected: true,
		},
		{
			name:     "boolean: false",
			cell:     false,
			typ:      sheetah.ColumnTypeBool,
			expected: false,
		},
		{
			name:     "string: true",
			cell:     true,
			typ:      sheetah.ColumnTypeString,
			expected: "true",
		},
		{
			name:     "string: false",
			cell:     false,
			typ:      sheetah.ColumnTypeString,
			expected: "false",
		},
		{
			name:     "number: true",
			cell:     true,
			typ:      sheetah.ColumnTypeNumber,
			expected: int64(1),
		},
		{
			name:     "number: false",
			cell:     false,
			typ:      sheetah.ColumnTypeNumber,
			expected: int64(0),
		},
		{
			name:     "invalid type",
			cell:     true,
			typ:      sheetah.ColumnTypeTimestamp,
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cell.Value(tt.typ)
			AssertDiff(t, tt.expected, result)
		})
	}
}

func TestNilCell_Value(t *testing.T) {
	t.Parallel()

	result := sheetah.NilCell{}.Value(sheetah.ColumnTypeString)
	AssertDiff(t, nil, result)
}
