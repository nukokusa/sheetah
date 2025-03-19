package sheetah

import (
	"context"
	"fmt"
)

type ExportOption struct {
	SpreadsheetID string `help:"Spreadsheet ID" required:"" name:"id" env:"SHEETAH_SPREADSHEET_ID"`
	Format        string `help:"Export format [yaml,json]" default:"yaml" enum:"yaml,json"`
	Dir           string `help:"Export directory" type:"path" default:"."`
}

func (o *ExportOption) Validate() error {
	return nil
}

func (c *CLI) runExport(ctx context.Context, opt *ExportOption) error {
	var err error
	if err = opt.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	config, err := LoadConfig(c.Config)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fetcher, err := NewFetcher(ctx, c.Credential)
	if err != nil {
		return fmt.Errorf("failed to create fetcher: %w", err)
	}

	sheets, err := fetcher.FetchSheets(ctx, opt.SpreadsheetID, config.Sheets)
	if err != nil {
		return fmt.Errorf("failed to fetch sheets: %w", err)
	}

	var o Outputter
	switch opt.Format {
	case "yaml":
		o = &YAMLOutputter{}
	case "json":
		o = &JSONOutputter{}
	default:
		o = &YAMLOutputter{}
	}

	if err := o.Output(ctx, opt.Dir, sheets...); err != nil {
		return fmt.Errorf("failed to output sheets: %w", err)
	}
	return nil
}
