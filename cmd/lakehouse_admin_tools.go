package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	arg "github.com/alexflint/go-arg"
	obs_cmd "github.com/unhealme/lakehouse-admin-tools/cmd/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	obs_int "github.com/unhealme/lakehouse-admin-tools/internal/obs"
)

const Version = "0.0.2"

type OBSArguments struct {
	BatchRename *obs_cmd.BatchRenameArgs `arg:"subcommand:batch-rename"`

	Endpoint *string `arg:"-e,--endpoint,env:OBS_ENDPOINT" placeholder:"ENDPOINT"`
}

type Arguments struct {
	OBS *OBSArguments `arg:"subcommand:obs"`

	ConfigFile   *string `arg:"-c,--config,env:LHAT_CONFIG" help:"load config from FILE" placeholder:"FILE"`
	AccessKey    *string `arg:"-,--ak,env:HW_ACCESS_KEY" placeholder:"ACCESS_KEY"`
	SecretKey    *string `arg:"-,--sk,env:HW_SECRET_KEY" placeholder:"SECRET_KEY"`
	SessionToken *string `arg:"-,--token,env:HW_SECURITY_TOKEN" placeholder:"SECURITY_TOKEN"`
	Verbose      bool    `arg:"-v,--verbose" help:"enable debug logging"`
}

func (a Arguments) ToArgs() []any {
	var (
		args []any
		v    = reflect.ValueOf(a)
	)
	for _, f := range reflect.VisibleFields(reflect.TypeFor[Arguments]()) {
		if !strings.HasPrefix(f.Name, "_") {
			args = append(args, f.Name)
			args = append(args, fmt.Sprintf("%#v", v.FieldByName(f.Name)))
		}
	}
	return args
}

type compVer struct {
	name    string
	version string
}

func (*Arguments) Epilogue() string {
	var b strings.Builder
	fmt.Fprintln(&b, "Components:")
	for _, v := range []compVer{{"obs-batch-rename", obs_cmd.BatchRenameVersion}} {
		fmt.Fprintf(&b, "%s %s\n", v.name, v.version)
	}
	return strings.TrimRight(b.String(), "\n")
}

func (*Arguments) Version() string {
	return fmt.Sprintf("%s %s (%s-%s)", os.Args[0], Version, runtime.GOOS, runtime.GOARCH)
}

var logger = internal.DefaultLogger()

func main() {
	args := &Arguments{}
	arg.MustParse(args)
	if args.Verbose {
		logger.Level = internal.LogLevelDebug
	}
	logger.Debug("arguments parsed.", logger.Args(args.ToArgs()...))

	cfg := internal.GetConfig(logger, args.ConfigFile)
	if args.AccessKey != nil {
		cfg.AccessKey = *args.AccessKey
	}
	if args.SecretKey != nil {
		cfg.SecretKey = *args.SecretKey
	}
	if args.SessionToken != nil {
		cfg.SessionToken = args.SessionToken
	}

	switch {
	case args.OBS != nil:
		if args.OBS.Endpoint != nil {
			cfg.OBS.Endpoint = *args.OBS.Endpoint
		}

		obsClient, err := obs_int.New(cfg.OBS.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)
		if err != nil {
			logger.Fatal("unable to create OBS client.", logger.Args("error", err))
		}
		defer obsClient.Close()

		switch {
		case args.OBS.BatchRename != nil:
			if !strings.HasSuffix(args.OBS.BatchRename.Path, "/") {
				args.OBS.BatchRename.Path += "/"
			}
			obs_cmd.BatchRename(obsClient, args.OBS.BatchRename)
		}
	}
}
