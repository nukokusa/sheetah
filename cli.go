package sheetah

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Config     string `help:"Load configuration from FILE" short:"c" type:"existingfile" default:"sheetah.yaml"`
	Credential string `help:"JSON credential file for access to spreadsheet" type:"existingfile" default:"credential.json"`

	Validate *ValidateOption `cmd:"" help:"Validate configuration file."`
	Export   *ExportOption   `cmd:"" help:"Export sheets to files."`

	Debug   bool             `help:"Enable debug mode" hidden:""`
	Version kong.VersionFlag `short:"v" help:"Show version."`
}

type ExportCLI struct{}

func (c *CLI) Run(ctx context.Context) error {
	k := kong.Parse(c, kong.Vars{"version": fmt.Sprintf("sheetah %s", Version)})
	if c.Debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	var err error
	switch k.Command() {
	case "validate":
		err = c.runValidate(ctx, c.Validate)
	case "export":
		err = c.runExport(ctx, c.Export)
	default:
		err = fmt.Errorf("unknown command: %s", k.Command())
	}
	return err
}
