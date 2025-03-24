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

var retryPolicy = retry.Policy{
	MinDelay: time.Second,
	MaxDelay: 100 * time.Second,
	MaxCount: 10,
}

type Fetcher struct {
	ss *sheets.Service
}

func NewFetcher(ctx context.Context, credentials string) (*Fetcher, error) {
	sheetsService, err := sheets.NewService(ctx,
		option.WithCredentialsFile(credentials),
		option.WithScopes(sheets.SpreadsheetsScope),
	)
	if err != nil {
		return nil, err
	}

	return &Fetcher{
		ss: sheetsService,
	}, nil
}

func (f *Fetcher) FetchSheets(ctx context.Context, spreadsheetID string, configs []*SheetConfig) ([]*Sheet, error) {
	retrier := retryPolicy.Start(ctx)

	req := f.ss.Spreadsheets.Get(spreadsheetID).Fields("properties.timeZone")
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

	batchReq := f.ss.Spreadsheets.Values.BatchGet(spreadsheetID)
	ranges := lo.Map(configs, func(config *SheetConfig, _ int) string {
		if config.Range == "" {
			return config.Name
		}
		return config.Name + "!" + config.Range
	})
	batchReq.Ranges(ranges...)
	batchReq.ValueRenderOption("UNFORMATTED_VALUE")
	batchReq.DateTimeRenderOption("FORMATTED_STRING")
	var batchResp *sheets.BatchGetValuesResponse
	for retrier.Continue() {
		var err error
		batchResp, err = batchReq.Context(ctx).Do()
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

	return NewSheets(spreadsheetID, configs, resp, batchResp.ValueRanges)
}
