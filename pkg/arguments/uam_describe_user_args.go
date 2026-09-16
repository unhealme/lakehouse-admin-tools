package arguments

import (
	"errors"
	"strings"

	"github.com/unhealme/lakehouse-admin-tools/internal/clients/uam"
)

type UamDescribeUserArgs struct {
	Users      []string                   `arg:"positional" placeholder:"USER"`
	InputFile  string                     `arg:"-i,--" help:"read user input from FILE" placeholder:"FILE"`
	OutputFile string                     `arg:"-o,--" help:"write result to FILE instead of stdout" placeholder:"FILE"`
	Format     UamDescribeUserPrintFormat `arg:"-f,--format" default:"default" help:"output format" placeholder:"{default,csv}"`
	NoHeader   bool                       `arg:"-,--no-header" help:"do not print header for csv output format"`
	Unsafe     bool                       `arg:"-,--unsafe" help:"do not escape USER"`

	BaseDn    string         `arg:"-"`
	GroupBase string         `arg:"-"`
	UamClient *uam.UamClient `arg:"-"`
}

type UamDescribeUserPrintFormat int

const (
	UamDescribeUsePrintFormatDefault UamDescribeUserPrintFormat = iota + 1
	UamDescribeUsePrintFormatCsv
)

func (f *UamDescribeUserPrintFormat) UnmarshalText(buf []byte) (e error) {
	switch fmt := string(buf); strings.ToLower(strings.TrimSpace(fmt)) {
	case "default":
		*f = UamDescribeUsePrintFormatDefault
	case "csv":
		*f = UamDescribeUsePrintFormatCsv
	default:
		return errors.New("invalid output format: " + fmt)
	}
	return
}
