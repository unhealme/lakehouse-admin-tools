package internal

import (
	"errors"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type EmptyType struct{}

var Empty EmptyType

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalText(buf []byte) error {
	dur, err := time.ParseDuration(string(buf))
	if err != nil {
		return err
	}
	*d = Duration{dur}
	return nil
}

type DateRangeKind int

const (
	DateRangeConstraint DateRangeKind = iota + 1
	DateRangePattern
	DateRangeRegex
	DateRangeArray
)

type DateRangeParsed struct {
	Kind                   DateRangeKind
	Start, End             *time.Time
	Format, Pattern, Regex string
	Array                  []string
}

func (d *DateRangeParsed) UnmarshalYAML(node ast.Node) error {
	var dateArray []string
	dec := yaml.NewDecoder(node)
	if err := dec.Decode(&dateArray); err == nil {
		d.Kind = DateRangeArray
		d.Array = dateArray
		return nil
	}

	var dateRange struct {
		Start, End, Format,
		Pattern,
		Regex string
	}
	if err := dec.Decode(&dateRange); err != nil {
		return err
	}

	switch {
	case (dateRange.Start != "" || dateRange.End != "") && dateRange.Format != "":
		var start, end *time.Time
		if dateRange.Start != "" {
			start = &time.Time{}
			if err := ParseStrftime(dateRange.Start, dateRange.Format, start); err != nil {
				return err
			}
		}
		if dateRange.End != "" {
			end = &time.Time{}
			if err := ParseStrftime(dateRange.End, dateRange.Format, end); err != nil {
				return err
			}
		}
		*d = DateRangeParsed{DateRangeConstraint, start, end, dateRange.Format, "", "", nil}
	case dateRange.Pattern != "":
		*d = DateRangeParsed{DateRangePattern, nil, nil, "", dateRange.Pattern, "", nil}
	case dateRange.Regex != "":
		*d = DateRangeParsed{DateRangeRegex, nil, nil, "", "", dateRange.Regex, nil}
	default:
		return errors.New("date range field combinations are not correct.")
	}
	return nil
}
