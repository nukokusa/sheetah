# sheetah

sheetah is a tool for exporting data from Google Sheets.

## Usage

```
Usage: sheetah <command> [flags]

Flags:
  -h, --help                            Show context-sensitive help.
  -c, --config="sheetah.yaml"           Load configuration from FILE
      --credential="credential.json"    JSON credential file for access to spreadsheet
  -v, --version                         Show version.

Commands:
  validate [flags]
    Validate configuration file.

  export --id=STRING [flags]
    Export sheets to files.
```

```
Usage: sheetah export --id=STRING [flags]

Export sheets to files.

Flags:
      --id=STRING                  Spreadsheet ID ($SHEETAH_SPREADSHEET_ID)
      --format="yaml"              Export format [yaml,json]
      --dir="."                    Export directory
      --sheets=SHEETS,...          Sheet names (matching a SheetConfig name) to
                                   export; exports every configured sheet when
                                   omitted
```

Each configured sheet is written to `<dir>/<name>.yaml` or `<dir>/<name>.json`,
depending on `--format`.

`--sheets` restricts which configured sheets are exported, by their `name` in
the configuration file (e.g. `--sheets=weapon,item`). When omitted, every
sheet in the configuration is exported. Naming a sheet that isn't in the
configuration is an error.

## Configurations

Configuration file is YAML format. Describe the structure of the table.

```yaml
sheets:
  - name: weapon
    range: A1:D10
    id_column: id
    columns:
      - name: id
        type: number
      - name: name
        type: string
      - name: damage
        type: number
      - name: release_date
        type: timestamp
        format: "2006-01-02"
  - name: item
    columns:
      - name: id
        type: number
      - name: name
        type: string
      - name: consumable
        type: boolean
```

The `type` specifies the data type of the column. The following types can be used:
- number
- string
- boolean
- timestamp

When using `timestamp`, if the timezone is not specified in the sheet value, the spreadsheet settings will be used.

If the sheet value does not match the type, it will not be output.

- `id_column` optionally names one of `columns` as the row's identifier (at
  most one per sheet). A row whose `id_column` value is the zero value for
  its type (`0`, `""`, `false`, a zero timestamp, or a missing/unparseable
  value) is excluded from the output. Must match one of `columns`' `name`s.
  When set, output rows are sorted in ascending order by this column's value.
- `format` optionally specifies a Go reference-time layout (e.g.
  `2006-01-02`, or `time.RFC3339`'s layout) used to render a `timestamp`
  column's value in the output. Only valid when `type` is `timestamp`;
  defaults to an RFC 3339 string when omitted.

## Author

Copyright (c) 2025 Daisuke Nagashima

## LICENSE

MIT
