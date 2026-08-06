package args

import (
	"github.com/unhealme/lakehouse-admin-tools/internal/fim"
)

type FimResetUserPasswordArgs struct {
	Users       []string `arg:"positional" placeholder:"USER"`
	DefaultPass string   `arg:"-d,--default-pass,env:FIM_DEFAULT_PASSWORD" placeholder:"FIM_DEFAULT_PASSWORD"`
	NoProg      bool     `arg:"-,--no-progress" help:"disable progress bar"`

	FimClient *fim.FimClient `arg:"-"`
	LoginUser string         `arg:"-"`
	LoginPass string         `arg:"-"`
}
