package sheetah

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
)

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = cerr
		}
	}()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

type Config struct {
	Sheets []*SheetConfig `yaml:"sheets"`
}

func (c *Config) Validate() error {
	if len(c.Sheets) == 0 {
		return errors.New("sheets is empty")
	}
	for _, sc := range c.Sheets {
		if err := sc.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type SheetConfig struct {
	Name    string          `yaml:"name"`
	Range   string          `yaml:"range,omitempty"`
	Columns []*ColumnConfig `yaml:"columns"`
}

var (
	a1Regex   = regexp.MustCompile(`^([A-Z]+[0-9]+)(:[A-Z]+[0-9]+)?$`)
	r1c1Regex = regexp.MustCompile(`^(R[0-9]+C[0-9]+)(:R[0-9]+C[0-9]+)?$`)
)

func (sc *SheetConfig) Validate() error {
	if sc.Name == "" {
		return errors.New("name is empty")
	}
	if sc.Range != "" {
		if !a1Regex.MatchString(sc.Range) && !r1c1Regex.MatchString(sc.Range) {
			return fmt.Errorf("invalid range format: %s", sc.Range)
		}
	}
	if len(sc.Columns) == 0 {
		return errors.New("columns is empty")
	}
	for _, cc := range sc.Columns {
		if err := cc.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type ColumnConfig struct {
	Name string     `yaml:"name"`
	Type ColumnType `yaml:"type"`
}

func (cc *ColumnConfig) Validate() error {
	if cc.Name == "" {
		return errors.New("column name is required")
	}
	if cc.Type == "" {
		return errors.New("column type is required")
	}
	if err := cc.Type.Validate(); err != nil {
		return err
	}

	return nil
}

type ColumnType string

const (
	ColumnTypeString    ColumnType = "string"
	ColumnTypeNumber    ColumnType = "number"
	ColumnTypeBool      ColumnType = "boolean"
	ColumnTypeTimestamp ColumnType = "timestamp"
)

func (ct ColumnType) Validate() error {
	switch ct {
	case ColumnTypeString, ColumnTypeNumber, ColumnTypeBool, ColumnTypeTimestamp:
		return nil
	default:
		return fmt.Errorf("not supported column type: %s", ct)
	}
}
