package sheetah_test

import (
	"encoding/json"
	"testing"

	"github.com/nukokusa/sheetah"
	"google.golang.org/api/sheets/v4"
)

// newTestSheets is a small helper wrapping NewSheets for the common case of
// a single configured sheet, given as a raw [][]any grid (header row plus
// data rows) rather than a full BatchGet response.
func newTestSheets(t *testing.T, config *sheetah.SheetConfig, grid [][]any) []*sheetah.Sheet {
	t.Helper()

	spreadsheet := &sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{}}
	valueRanges := []*sheets.ValueRange{
		{Range: config.Name + "!A1", Values: grid},
	}

	shs, err := sheetah.NewSheets("spreadsheet-id", []*sheetah.SheetConfig{config}, spreadsheet, valueRanges)
	if err != nil {
		t.Fatal(err)
	}
	return shs
}

func marshalRows(t *testing.T, sheet *sheetah.Sheet) []map[string]any {
	t.Helper()

	b, err := sheet.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(b, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestSheet_IDColumn_ExcludesZeroValueRows(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name:     "item",
		IDColumn: "id",
		Columns: []*sheetah.ColumnConfig{
			{Name: "id", Type: sheetah.ColumnTypeNumber},
			{Name: "name", Type: sheetah.ColumnTypeString},
		},
	}

	grid := [][]any{
		{"id", "name"},         // header row
		{float64(1), "Potion"}, // kept: non-zero id
		{float64(0), "Ghost"},  // excluded: zero id
		{nil, "No ID"},         // excluded: missing id
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %v", len(rows), rows)
	}
	if rows[0]["name"] != "Potion" {
		t.Errorf("expected the surviving row to be Potion, got %v", rows[0])
	}
}

func TestSheet_IDColumn_StringZeroValue(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name:     "item",
		IDColumn: "code",
		Columns: []*sheetah.ColumnConfig{
			{Name: "code", Type: sheetah.ColumnTypeString},
		},
	}

	grid := [][]any{
		{"code"},
		{"A1"}, // kept
		{""},   // excluded: empty string is the zero value
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %v", len(rows), rows)
	}
	if rows[0]["code"] != "A1" {
		t.Errorf("expected the surviving row to be A1, got %v", rows[0])
	}
}

func TestSheet_IDColumn_SortsAscendingByNumber(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name:     "item",
		IDColumn: "id",
		Columns: []*sheetah.ColumnConfig{
			{Name: "id", Type: sheetah.ColumnTypeNumber},
			{Name: "name", Type: sheetah.ColumnTypeString},
		},
	}

	grid := [][]any{
		{"id", "name"},
		{float64(3), "Charm"},
		{float64(1), "Potion"},
		{float64(2.5), "Elixir"}, // a fractional id sorts between 2 and 3
		{float64(2), "Hi-Potion"},
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	wantOrder := []string{"Potion", "Hi-Potion", "Elixir", "Charm"}
	if len(rows) != len(wantOrder) {
		t.Fatalf("expected %d rows, got %d: %v", len(wantOrder), len(rows), rows)
	}
	for i, want := range wantOrder {
		if rows[i]["name"] != want {
			t.Errorf("row %d: name = %v, want %s", i, rows[i]["name"], want)
		}
	}
}

func TestSheet_IDColumn_SortsAscendingByString(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name:     "item",
		IDColumn: "code",
		Columns: []*sheetah.ColumnConfig{
			{Name: "code", Type: sheetah.ColumnTypeString},
		},
	}

	grid := [][]any{
		{"code"},
		{"B1"},
		{"A1"},
		{"C1"},
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	wantOrder := []string{"A1", "B1", "C1"}
	if len(rows) != len(wantOrder) {
		t.Fatalf("expected %d rows, got %d: %v", len(wantOrder), len(rows), rows)
	}
	for i, want := range wantOrder {
		if rows[i]["code"] != want {
			t.Errorf("row %d: code = %v, want %s", i, rows[i]["code"], want)
		}
	}
}

func TestSheet_IDColumn_SortsAscendingByTimestamp(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name:     "item",
		IDColumn: "created_at",
		Columns: []*sheetah.ColumnConfig{
			{Name: "created_at", Type: sheetah.ColumnTypeTimestamp},
			{Name: "name", Type: sheetah.ColumnTypeString},
		},
	}

	grid := [][]any{
		{"created_at", "name"},
		{"2024-03-01T00:00:00Z", "third"},
		{"2024-01-01T00:00:00Z", "first"},
		{"2024-02-01T00:00:00Z", "second"},
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	wantOrder := []string{"first", "second", "third"}
	if len(rows) != len(wantOrder) {
		t.Fatalf("expected %d rows, got %d: %v", len(wantOrder), len(rows), rows)
	}
	for i, want := range wantOrder {
		if rows[i]["name"] != want {
			t.Errorf("row %d: name = %v, want %s", i, rows[i]["name"], want)
		}
	}
}

func TestSheet_TimestampColumn_CustomFormat(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name: "item",
		Columns: []*sheetah.ColumnConfig{
			{Name: "id", Type: sheetah.ColumnTypeNumber},
			{Name: "released_at", Type: sheetah.ColumnTypeTimestamp, Format: "2006-01-02"},
		},
	}

	grid := [][]any{
		{"id", "released_at"},
		{float64(1), "2024-01-02T15:04:05Z"},
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %v", len(rows), rows)
	}
	if rows[0]["released_at"] != "2024-01-02" {
		t.Errorf("released_at = %v, want the custom-formatted string 2024-01-02", rows[0]["released_at"])
	}
}

func TestSheet_TimestampColumn_DefaultFormat(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name: "item",
		Columns: []*sheetah.ColumnConfig{
			{Name: "released_at", Type: sheetah.ColumnTypeTimestamp},
		},
	}

	grid := [][]any{
		{"released_at"},
		{"2024-01-02T15:04:05Z"},
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	if rows[0]["released_at"] != "2024-01-02T15:04:05Z" {
		t.Errorf("released_at = %v, want the default RFC3339 rendering", rows[0]["released_at"])
	}
}

func TestSheet_IDColumn_TimestampWithFormat_ChecksRawZeroness(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name:     "item",
		IDColumn: "created_at",
		Columns: []*sheetah.ColumnConfig{
			// A format that would never render as an empty string, to make
			// sure the zero check happens against the underlying
			// time.Time, not against the already-formatted string.
			{Name: "created_at", Type: sheetah.ColumnTypeTimestamp, Format: "2006"},
		},
	}

	grid := [][]any{
		{"created_at"},
		{"2024-01-02T00:00:00Z"}, // kept: not a zero time
		{"0001-01-01T00:00:00Z"}, // excluded: time.Time's zero value
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %v", len(rows), rows)
	}
	if rows[0]["created_at"] != "2024" {
		t.Errorf("created_at = %v, want 2024", rows[0]["created_at"])
	}
}

func TestSheet_NoIDColumn_KeepsAllRows(t *testing.T) {
	t.Parallel()

	config := &sheetah.SheetConfig{
		Name: "item",
		Columns: []*sheetah.ColumnConfig{
			{Name: "id", Type: sheetah.ColumnTypeNumber},
		},
	}

	grid := [][]any{
		{"id"},
		{float64(1)},
		{float64(0)},
	}

	shs := newTestSheets(t, config, grid)
	rows := marshalRows(t, shs[0])

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (no filtering without id_column), got %d: %v", len(rows), rows)
	}
}
