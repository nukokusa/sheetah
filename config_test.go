package sheetah_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nukokusa/sheetah"
)

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sheetah.yaml")
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConfig_FilterSheets(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, `
sheets:
  - name: weapon
    columns:
      - name: id
        type: number
  - name: item
    columns:
      - name: id
        type: number
  - name: armor
    columns:
      - name: id
        type: number
`)
	config, err := sheetah.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("no names means no filtering", func(t *testing.T) {
		t.Parallel()
		filtered, err := config.FilterSheets(nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(filtered) != 3 {
			t.Fatalf("expected 3 sheets, got %d", len(filtered))
		}
	})

	t.Run("filters down to the requested names, keeping config order", func(t *testing.T) {
		t.Parallel()
		filtered, err := config.FilterSheets([]string{"armor", "weapon"})
		if err != nil {
			t.Fatal(err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 sheets, got %d", len(filtered))
		}
		if filtered[0].Name != "weapon" || filtered[1].Name != "armor" {
			t.Errorf("expected [weapon armor] in config order, got [%s %s]", filtered[0].Name, filtered[1].Name)
		}
	})

	t.Run("unknown name is an error", func(t *testing.T) {
		t.Parallel()
		if _, err := config.FilterSheets([]string{"weapon", "does-not-exist"}); err == nil {
			t.Fatal("expected an error for an unknown sheet name")
		}
	})
}

func TestSheetConfig_IDColumn_Validate(t *testing.T) {
	t.Parallel()

	t.Run("id_column matching a column is valid", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, `
sheets:
  - name: item
    id_column: id
    columns:
      - name: id
        type: number
      - name: name
        type: string
`)
		if _, err := sheetah.LoadConfig(path); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("id_column not matching any column is an error", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, `
sheets:
  - name: item
    id_column: does-not-exist
    columns:
      - name: id
        type: number
`)
		if _, err := sheetah.LoadConfig(path); err == nil {
			t.Fatal("expected an error for an id_column that isn't in columns")
		}
	})

	t.Run("id_column is optional", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, `
sheets:
  - name: item
    columns:
      - name: id
        type: number
`)
		if _, err := sheetah.LoadConfig(path); err != nil {
			t.Fatal(err)
		}
	})
}

func TestColumnConfig_Format_Validate(t *testing.T) {
	t.Parallel()

	t.Run("format on a timestamp column is valid", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, `
sheets:
  - name: item
    columns:
      - name: released_at
        type: timestamp
        format: "2006-01-02"
`)
		if _, err := sheetah.LoadConfig(path); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("format on a non-timestamp column is an error", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, `
sheets:
  - name: item
    columns:
      - name: name
        type: string
        format: "2006-01-02"
`)
		if _, err := sheetah.LoadConfig(path); err == nil {
			t.Fatal("expected an error for format on a non-timestamp column")
		}
	})

	t.Run("format is optional", func(t *testing.T) {
		t.Parallel()
		path := writeConfig(t, `
sheets:
  - name: item
    columns:
      - name: released_at
        type: timestamp
`)
		if _, err := sheetah.LoadConfig(path); err != nil {
			t.Fatal(err)
		}
	})
}
