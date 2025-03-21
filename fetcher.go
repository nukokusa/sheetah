package sheetah

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/samber/lo"
	"github.com/shogo82148/go-retry/v2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Fetcher interface {
	FetchSheets(ctx context.Context, spreadsheetID string, configs []*SheetConfig) ([]*Sheet, error)
}

type fetcher struct {
	ss *sheets.Service
}

func NewFetcher(ctx context.Context, credentials string) (Fetcher, error) {
	sheetsService, err := sheets.NewService(ctx,
		option.WithCredentialsFile(credentials),
		option.WithScopes(sheets.SpreadsheetsScope),
	)
	if err != nil {
		return nil, err
	}

	return &fetcher{
		ss: sheetsService,
	}, nil
}

var (
	retryPolicy = retry.Policy{
		MinDelay: time.Second,
		MaxDelay: 100 * time.Second,
		MaxCount: 10,
	}
	fields = []googleapi.Field{
		"properties.timeZone",
		"sheets.properties.title",
		"sheets.data.rowData.values.userEnteredFormat.numberFormat.type",
		"sheets.data.rowData.values.formattedValue",
		"sheets.data.rowData.values.effectiveValue",
	}
)

func (f *fetcher) FetchSheets(ctx context.Context, spreadsheetID string, configs []*SheetConfig) ([]*Sheet, error) {
	req := f.ss.Spreadsheets.Get(spreadsheetID).IncludeGridData(true)
	ranges := lo.Map(configs, func(config *SheetConfig, _ int) string {
		return config.Name
	})
	req.Ranges(ranges...)
	req.Fields(fields...)
	retrier := retryPolicy.Start(ctx)
	var resp *sheets.Spreadsheet
	for retrier.Continue() {
		var err error
		resp, err = req.Context(ctx).Do()
		if err == nil {
			break
		}
		if apiError, ok := err.(*googleapi.Error); ok {
			if apiError.Code == http.StatusTooManyRequests {
				slog.Warn("Rate limit exceeded, retrying...", "error", err)
				continue
			}
			return nil, err
		}
		return nil, err
	}
	return NewSheets(spreadsheetID, configs, resp)
}
