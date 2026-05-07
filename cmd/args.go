package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// OBS Batch Rename
type BatchRenameArgs struct {
	Path        string `arg:"positional,required" placeholder:"PATH"`
	Prefix      string `arg:"positional,required" placeholder:"PREFIX"`
	Concurrency int    `arg:"-j,--" default:"2" help:"max job concurrency" placeholder:"NUM"`
	DirOnly     bool   `arg:"-d,--dir-only" help:"only rename directories"`
	DryRun      bool   `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg      bool   `arg:"-,--no-progress" help:"disable progress bar"`
}

// OBS Batch Set Storage Class
type BatchSetStorageClassArgs struct {
	InputFiles  []string `arg:"positional,required" placeholder:"FILE"`
	Concurrency int      `arg:"-j,--" default:"2" help:"max job concurrency" placeholder:"NUM"`
	DryRun      bool     `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg      bool     `arg:"-,--no-progress" help:"disable progress bar"`
}

type OBSArguments struct {
	BatchRename          *BatchRenameArgs          `arg:"subcommand:batch-rename"`
	BatchSetStorageClass *BatchSetStorageClassArgs `arg:"subcommand:batch-set-storage-class"`

	Endpoint string `arg:"-e,--endpoint,env:OBS_ENDPOINT" placeholder:"ENDPOINT"`
}

type Arguments struct {
	OBS *OBSArguments `arg:"subcommand:obs"`

	ConfigFile   string `arg:"-c,--config,env:LHAT_CONFIG" help:"load config from FILE" placeholder:"FILE"`
	AccessKey    string `arg:"-,--ak,env:HW_ACCESS_KEY" placeholder:"ACCESS_KEY"`
	SecretKey    string `arg:"-,--sk,env:HW_SECRET_KEY" placeholder:"SECRET_KEY"`
	SessionToken string `arg:"-,--token,env:HW_SECURITY_TOKEN" placeholder:"SECURITY_TOKEN"`
	Verbose      bool   `arg:"-v,--verbose" help:"enable debug logging"`
}

func (Arguments) Epilogue() string {
	var b strings.Builder
	fmt.Fprintln(&b, "Components:")
	for comp, ver := range compVer {
		fmt.Fprintf(&b, "  %s %s\n", comp, ver)
	}
	return strings.TrimRight(b.String(), "\n")
}

func (Arguments) Version() string {
	return fmt.Sprintf("%s %s (%s-%s)", os.Args[0], Version, runtime.GOOS, runtime.GOARCH)
}
