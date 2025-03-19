package sheetah_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nukokusa/sheetah"
	"google.golang.org/api/sheets/v4"
)

func AssertDiff(t *testing.T, expected, actual any, opts ...cmp.Option) {
	t.Helper()

	if diff := cmp.Diff(expected, actual, opts...); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestCell_number(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cell     *sheetah.Cell
		expected any
	}{
		{
			name: "FormatType: Unspecified",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						NumberFormat: &sheets.NumberFormat{
							Type: string(sheetah.CellNumberFormatTypeUnspecified),
						},
					},
					EffectiveValue: &sheets.ExtendedValue{
						NumberValue: func() *float64 {
							f := float64(12.3)
							return &f
						}(),
					},
				},
			},
			expected: float64(12.3),
		},
		{
			name: "FormatType: Unspecified, integer",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						NumberFormat: &sheets.NumberFormat{
							Type: string(sheetah.CellNumberFormatTypeUnspecified),
						},
					},
					EffectiveValue: &sheets.ExtendedValue{
						NumberValue: func() *float64 {
							f := float64(12.0)
							return &f
						}(),
					},
				},
			},
			expected: int64(12),
		},
		{
			name: "FormatType: Percent",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						NumberFormat: &sheets.NumberFormat{
							Type: string(sheetah.CellNumberFormatTypePercent),
						},
					},
					EffectiveValue: &sheets.ExtendedValue{
						NumberValue: func() *float64 {
							f := float64(0.123)
							return &f
						}(),
					},
				},
			},
			expected: float64(12.3),
		},
		{
			name: "FormatType: Date",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						NumberFormat: &sheets.NumberFormat{
							Type: string(sheetah.CellNumberFormatTypeDate),
						},
					},
					EffectiveValue: &sheets.ExtendedValue{
						NumberValue: func() *float64 {
							f := float64(36526) // 2000/01/01
							return &f
						}(),
					},
				},
			},
			expected: nil,
		},
		{
			name: "StringValue",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "12.3"
							return &s
						}(),
					},
				},
			},
			expected: float64(12.3),
		},
		{
			name: "StringValue, integer",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "12.0"
							return &s
						}(),
					},
				},
			},
			expected: int64(12),
		},
		{
			name: "failed to parse number",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						ErrorValue: &sheets.ErrorValue{
							Message: "dummy message",
							Type:    "ERROR",
						},
					},
				},
			},
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sheetah.CellNumber(tt.cell)
			AssertDiff(t, tt.expected, result)
		})
	}
}

func TestCell_bool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cell     *sheetah.Cell
		expected any
	}{
		{
			name: "BoolValue: true",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						BoolValue: func() *bool {
							b := true
							return &b
						}(),
					},
				},
			},
			expected: true,
		},
		{
			name: "BoolValue: false",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						BoolValue: func() *bool {
							b := false
							return &b
						}(),
					},
				},
			},
			expected: false,
		},
		{
			name: "NumberValue: zero",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						NumberValue: func() *float64 {
							f := float64(0)
							return &f
						}(),
					},
				},
			},
			expected: false,
		},
		{
			name: "NumberValue: non-zero",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						NumberValue: func() *float64 {
							f := float64(1)
							return &f
						}(),
					},
				},
			},
			expected: true,
		},
		{
			name: "StringValue: TRUE",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "TRUE"
							return &s
						}(),
					},
				},
			},
			expected: true,
		},
		{
			name: "StringValue: FALSE",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "FALSE"
							return &s
						}(),
					},
				},
			},
			expected: false,
		},
		{
			name: "failed to parse boolean",
			cell: &sheetah.Cell{
				CellData: &sheets.CellData{
					EffectiveValue: &sheets.ExtendedValue{
						ErrorValue: &sheets.ErrorValue{
							Message: "dummy message",
							Type:    "ERROR",
						},
					},
				},
			},
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sheetah.CellBool(tt.cell)
			AssertDiff(t, tt.expected, result)
		})
	}
}
