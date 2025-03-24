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

## Configurations

Configuration file is YAML format. Describe the structure of the table.

```yaml
sheets:
  - name: weapon
    range: A1:D10
    columns:
      - name: id
        type: number
      - name: name
        type: string
      - name: damage
        type: number
      - name: release_date
        type: timestamp
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

## Author

Copyright (c) 2025 Daisuke Nagashima

## LICENSE

MIT
