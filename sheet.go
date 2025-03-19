package sheetah

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/samber/lo"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	SpreadsheetID string
	Config        *SheetConfig
	Columns       []*sheets.CellData
	Rows          [][]*Cell
}

func NewSheet(spreadsheetID string, config *SheetConfig, spreadsheet *sheets.Spreadsheet) (*Sheet, error) {
	shs, err := NewSheets(spreadsheetID, []*SheetConfig{config}, spreadsheet)
	if err != nil {
		return nil, err
	}
	return shs[0], nil
}

func NewSheets(spreadsheetID string, configs []*SheetConfig, spreadsheet *sheets.Spreadsheet) ([]*Sheet, error) {
	if len(spreadsheet.Sheets) == 0 {
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
	sheetsByName := lo.SliceToMap(spreadsheet.Sheets, func(s *sheets.Sheet) (string, *sheets.Sheet) {
		return s.Properties.Title, s
	})

	isColumnRow := func(row *sheets.RowData, config *SheetConfig) bool {
		columnMap := lo.SliceToMap(config.Columns, func(c *ColumnConfig) (string, ColumnType) {
			return c.Name, c.Type
		})
		_, exist := lo.Find(row.Values, func(cell *sheets.CellData) bool {
			_, ok := columnMap[cell.FormattedValue]
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
		data := sheet.Data
		sort.Slice(data, func(i, j int) bool {
			return data[i].StartRow < data[j].StartRow
		})

		var columns []*sheets.CellData
		rows := [][]*Cell{}
		for _, d := range data {
			for _, row := range d.RowData {
				if columns == nil {
					if isColumnRow(row, config) {
						columns = row.Values
					}
					continue
				}
				rows = append(rows, lo.Map(row.Values, func(cell *sheets.CellData, _ int) *Cell {
					return &Cell{cell, loc}
				}))
			}
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
	for i, cell := range s.Columns {
		headers[i] = cell.FormattedValue
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
