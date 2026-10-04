package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	arg "github.com/alexflint/go-arg"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/obs"
)

type Arguments struct {
	InputFile  string `arg:"positional" default:"-" placeholder:"FILE" help:"read input from FILE. can be '-' to use stdin"`
	OutputFile string `arg:"-o,--" placeholder:"FILE" help:"write output to FILE. can be unset or '-' to use stdout"`
}

var obsClient = sync.OnceValue(func() *obs.ObsClient {
	endpoint := internal.GetEnv("OBS_ENDPOINT", "obs.ap-southeast-4.myhuaweicloud.com")
	c, err := obs.NewClient(
		endpoint,
		os.Getenv("OBS_ACCESS_KEY"),
		os.Getenv("OBS_SECRET_KEY"),
		os.Getenv("OBS_ACCESS_TOKEN"),
	)
	if err != nil {
		panic(err)
	}
	return c
})

func errOpenFile(context, name string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open file for %s: %s, error: %v\n", context, name, err)
		os.Exit(1)
	}
}

func main() {
	var args Arguments
	arg.MustParse(&args)

	var (
		in  io.ReadCloser  = os.Stdin
		out io.WriteCloser = os.Stdout

		wg sync.WaitGroup
	)

	if strings.HasPrefix(args.InputFile, "obs://") {
		obsPath, err := obs.PathFromURI(args.InputFile)
		if err != nil {
			errOpenFile("reading", args.InputFile, err)
		}
		if in, err = obsClient().ReadFile(*obsPath); err != nil {
			errOpenFile("reading", args.InputFile, err)
		}
		defer in.Close()
	} else if args.InputFile != "" && args.InputFile != "-" {
		var err error
		if in, err = os.Open(args.InputFile); err != nil {
			errOpenFile("reading", args.InputFile, err)
		}
		defer in.Close()
	}

	if strings.HasPrefix(args.OutputFile, "obs://") {
		obsPath, err := obs.PathFromURI(args.OutputFile)
		if err != nil {
			errOpenFile("writing", args.OutputFile, err)
		}

		r, w := io.Pipe()
		wg.Go(func() {
			if err := obsClient().WriteFile(*obsPath, r); err != nil {
				errOpenFile("writing", args.OutputFile, err)
			}
		})
		out = w
		defer r.Close()
	} else if args.OutputFile != "" && args.OutputFile != "-" {
		var err error
		if out, err = os.Create(args.OutputFile); err != nil {
			errOpenFile("writing", args.OutputFile, err)
		}
		defer out.Close()
	}

	pr, pw := io.Pipe()
	var decompressErr error
	go func() {
		var r io.Reader = in
		if !strings.HasPrefix(args.InputFile, "obs://") {
			r = bufio.NewReader(in)
		}
		decompressErr = decompress(r, pw)
		pw.Close()
	}()

	var parserErr error
	if strings.HasPrefix(args.OutputFile, "obs://") {
		parserErr = parser(pr, out)
		out.Close()
		pr.Close()
	} else {
		outBuf := bufio.NewWriter(out)
		parserErr = parser(pr, outBuf)
		outBuf.Flush()
		pr.Close()
	}
	wg.Wait()

	if decompressErr != nil {
		fmt.Fprintf(os.Stderr, "decompress error: %v\n", decompressErr)
	}
	if parserErr != nil {
		fmt.Fprintf(os.Stderr, "parser error: %v\n", parserErr)
	}
	if decompressErr != nil || parserErr != nil {
		os.Exit(1)
	}
}
