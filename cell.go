package sheetah

import (
	"log/slog"
	"math"
	"strconv"
	"time"

	"google.golang.org/api/sheets/v4"
)

type Cell struct {
	*sheets.CellData
	Loc *time.Location
}

func (c *Cell) Value(typ ColumnType) any {
	switch typ {
	case ColumnTypeString:
		return c.string()
	case ColumnTypeNumber:
		return c.number()
	case ColumnTypeBool:
		return c.bool()
	case ColumnTypeTimestamp:
		return c.timestamp()
	default:
		return c.CellData.FormattedValue
	}
}

func (c *Cell) string() string {
	return c.CellData.FormattedValue
}

func (c *Cell) number() any {
	floor := func(num float64) any {
		if math.Floor(num) == num {
			return int64(num)
		}
		return num
	}

	if c.CellData.EffectiveValue != nil {
		if c.CellData.EffectiveValue.NumberValue != nil {
			formatType := CellNumberFormatTypeUnspecified
			if c.CellData.UserEnteredFormat != nil && c.CellData.UserEnteredFormat.NumberFormat != nil {
				formatType = CellNumberFormatType(c.CellData.UserEnteredFormat.NumberFormat.Type)
			}

			switch formatType {
			case CellNumberFormatTypeUnspecified,
				CellNumberFormatTypeText,
				CellNumberFormatTypeNumber,
				CellNumberFormatTypeCurrency,
				CellNumberFormatTypeScientific:
				return floor(*c.CellData.EffectiveValue.NumberValue)
			case CellNumberFormatTypePercent:
				return floor(*c.CellData.EffectiveValue.NumberValue * 100)
			case CellNumberFormatTypeDate,
				CellNumberFormatTypeTime,
				CellNumberFormatTypeDateTime:
				// not supported format
			}
		}
		if c.CellData.EffectiveValue.StringValue != nil {
			if num, err := strconv.ParseFloat(*c.CellData.EffectiveValue.StringValue, 64); err == nil {
				return floor(num)
			}
		}
	}

	slog.Warn("Failed to parse number", "value", c.CellData.FormattedValue)
	return nil
}

func (c *Cell) bool() any {
	if c.CellData.EffectiveValue != nil {
		switch {
		case c.CellData.EffectiveValue.BoolValue != nil:
			return *c.CellData.EffectiveValue.BoolValue
		case c.CellData.EffectiveValue.NumberValue != nil:
			return *c.CellData.EffectiveValue.NumberValue != 0
		case c.CellData.EffectiveValue.StringValue != nil:
			if b, err := strconv.ParseBool(*c.CellData.EffectiveValue.StringValue); err == nil {
				return b
			}
		}
	}

	slog.Warn("Failed to parse boolean", "value", c.CellData.FormattedValue)
	return nil
}

func (c *Cell) timestamp() any {
	if c.CellData.EffectiveValue != nil {
		if c.CellData.EffectiveValue.NumberValue != nil {
			formatType := CellNumberFormatTypeUnspecified
			if c.CellData.UserEnteredFormat != nil && c.CellData.UserEnteredFormat.NumberFormat != nil {
				formatType = CellNumberFormatType(c.CellData.UserEnteredFormat.NumberFormat.Type)
			}

			switch formatType {
			case CellNumberFormatTypeUnspecified,
				CellNumberFormatTypeText:
				if t, err := ParseTimeByString(c.CellData.FormattedValue, c.Loc); err == nil {
					return t
				}
			case CellNumberFormatTypeDate,
				CellNumberFormatTypeDateTime:
				return ParseTimeBySerialNumber(*c.CellData.EffectiveValue.NumberValue, c.Loc)
			case CellNumberFormatTypeNumber,
				CellNumberFormatTypePercent,
				CellNumberFormatTypeCurrency,
				CellNumberFormatTypeScientific,
				CellNumberFormatTypeTime:
				// not supported format
			}
		}
		if c.CellData.EffectiveValue.StringValue != nil {
			if t, err := ParseTimeByString(c.CellData.FormattedValue, c.Loc); err == nil {
				return t
			}
		}
	}

	slog.Warn("Failed to parse timestamp", "value", c.CellData.FormattedValue)
	return nil
}

type CellNumberFormatType string

const (
	CellNumberFormatTypeUnspecified CellNumberFormatType = "NUMBER_FORMAT_TYPE_UNSPECIFIED"
	CellNumberFormatTypeText        CellNumberFormatType = "TEXT"
	CellNumberFormatTypeNumber      CellNumberFormatType = "NUMBER"
	CellNumberFormatTypePercent     CellNumberFormatType = "PERCENT"
	CellNumberFormatTypeCurrency    CellNumberFormatType = "CURRENCY"
	CellNumberFormatTypeDate        CellNumberFormatType = "DATE"
	CellNumberFormatTypeTime        CellNumberFormatType = "TIME"
	CellNumberFormatTypeDateTime    CellNumberFormatType = "DATE_TIME"
	CellNumberFormatTypeScientific  CellNumberFormatType = "SCIENTIFIC"
)
