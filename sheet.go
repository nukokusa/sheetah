package sheetah

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/samber/lo"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	SpreadsheetID string
	Config        *SheetConfig
	Columns       []string
	Rows          [][]Cell
}

func NewSheets(spreadsheetID string, configs []*SheetConfig, spreadsheet *sheets.Spreadsheet, valueRanges []*sheets.ValueRange) ([]*Sheet, error) {
	if len(valueRanges) == 0 {
		return nil, errors.New("sheet not found")
	}

	loc := time.UTC
	if spreadsheet.Properties.TimeZone != "" {
		var err error
		loc, err = time.LoadLocation(spreadsheet.Properties.TimeZone)
		if err != nil {
			return nil, err
		}
	}

	sheetsByName := lo.SliceToMap(valueRanges, func(vr *sheets.ValueRange) (string, *sheets.ValueRange) {
		parts := strings.SplitN(vr.Range, "!", 2)
		return parts[0], vr
	})

	formatString := func(cell any) string {
		switch _c := cell.(type) {
		case string:
			return _c
		case float64:
			return strconv.FormatFloat(_c, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(_c)
		default:
			return ""
		}
	}

	isColumnRow := func(row []any, config *SheetConfig) bool {
		columnMap := lo.SliceToMap(config.Columns, func(c *ColumnConfig) (string, ColumnType) {
			return c.Name, c.Type
		})
		_, exist := lo.Find(row, func(cell any) bool {
			_, ok := columnMap[formatString(cell)]
			return ok
		})
		return exist
	}

	var shs []*Sheet
	for _, config := range configs {
		sheet, ok := sheetsByName[config.Name]
		if !ok {
			return nil, fmt.Errorf("sheet not found: %s", config.Name)
		}

		var columns []string
		rows := [][]Cell{}
		for _, row := range sheet.Values {
			if columns == nil {
				if isColumnRow(row, config) {
					columns = lo.Map(row, func(cell any, _ int) string {
						return formatString(cell)
					})
				}
				continue
			}
			rows = append(rows, lo.Map(row, func(cell any, _ int) Cell {
				switch c := cell.(type) {
				case string:
					return NewStringCell(c, loc)
				case float64:
					return NumberCell(c)
				case bool:
					return BoolCell(c)
				default:
					return NilCell{}
				}
			}))
		}
		if columns == nil {
			return nil, fmt.Errorf("columns not found: %s", config.Name)
		}

		shs = append(shs, &Sheet{
			SpreadsheetID: spreadsheetID,
			Config:        config,
			Columns:       columns,
			Rows:          rows,
		})
	}

	return shs, nil
}

func (s *Sheet) Name() string {
	return s.Config.Name
}

func (s Sheet) marshal() []map[string]any {
	columnTypeMap := lo.SliceToMap(s.Config.Columns, func(c *ColumnConfig) (string, ColumnType) {
		return c.Name, c.Type
	})

	headers := make(map[int]string)
	for i, column := range s.Columns {
		headers[i] = column
	}

	var result []map[string]any
	for _, row := range s.Rows {
		rowMap := make(map[string]any)
		for i, cell := range row {
			name, ok := headers[i]
			if !ok {
				continue
			}
			column, ok := columnTypeMap[name]
			if !ok {
				continue
			}
			if value := cell.Value(column); value != nil {
				rowMap[name] = value
			}
		}
		result = append(result, rowMap)
	}
	return result
}

func (s Sheet) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.marshal())
}

func (s Sheet) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.marshal())
}
