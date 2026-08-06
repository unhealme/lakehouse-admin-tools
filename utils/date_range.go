package utils

import (
	"errors"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

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

type DateRangeRaw struct {
	Start, End, Format,
	Pattern,
	Regex string
}

func (d *DateRangeParsed) UnmarshalYAML(node ast.Node) (err error) {
	dec := yaml.NewDecoder(node)

	var dateArray []string
	if err = dec.Decode(&dateArray); err == nil {
		d.Kind = DateRangeArray
		d.Array = dateArray
		return
	}

	var dateRange DateRangeRaw
	if err = dec.Decode(&dateRange); err != nil {
		return
	}

	switch {
	case (dateRange.Start != "" || dateRange.End != "") && dateRange.Format != "":
		parsed := DateRangeParsed{
			Kind:   DateRangeConstraint,
			Format: dateRange.Format,
		}
		if dateRange.Start != "" {
			if d.Start, err = ParseStrftime(dateRange.Start, dateRange.Format); err != nil {
				return
			}
		}
		if dateRange.End != "" {
			if d.End, err = ParseStrftime(dateRange.End, dateRange.Format); err != nil {
				return
			}
		}
		*d = parsed
	case dateRange.Pattern != "":
		*d = DateRangeParsed{Kind: DateRangePattern, Pattern: dateRange.Pattern}
	case dateRange.Regex != "":
		*d = DateRangeParsed{Kind: DateRangeRegex, Regex: dateRange.Regex}
	default:
		err = errors.New("Date Range field combinations are not correct")
	}
	return
}
