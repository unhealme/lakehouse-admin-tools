package utils

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
)

type OutputWriter struct {
	csvWriter *csv.Writer
	csvFile   io.WriteCloser
	jsonFile  io.WriteCloser

	printCsvHeaderOnce sync.Once
}

func (w *OutputWriter) printCsvHeader() {
	w.printCsvHeaderOnce.Do(func() {
		if err := w.csvWriter.Write([]string{
			"Path",
			"Size",
			"SizeFormatted",
			"DirCount",
			"FileCount",
		}); err != nil {
			panic(err)
		}
	})
}

func (w *OutputWriter) Close() {
	if w.csvWriter != nil {
		w.csvWriter.Flush()
		w.csvFile.Close()
	}
	if w.jsonFile != nil {
		w.jsonFile.Close()
	}
}

func (w *OutputWriter) OpenCsvWriter(file string) (err error) {
	if w.csvFile, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644); err == nil {
		w.csvWriter = csv.NewWriter(w.csvFile)
	}
	return
}

func (w *OutputWriter) OpenJsonWriter(file string) (err error) {
	w.jsonFile, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	return
}

func (w *OutputWriter) Write(stats obs.ObsPathAnalyzed) {
	if w.csvWriter != nil {
		w.printCsvHeader()
		if err := w.csvWriter.Write([]string{
			stats.URI(),
			strconv.FormatInt(stats.Size, 10),
			internal.FormatSize(stats.Size),
			strconv.FormatInt(int64(stats.DirCount), 10),
			strconv.FormatInt(int64(stats.FileCount), 10),
		}); err != nil {
			panic(err)
		}
	}
	if w.jsonFile != nil {
		statsJson, _ := json.Marshal(stats)
		if _, err := w.jsonFile.Write(append(statsJson, '\n')); err != nil {
			panic(err)
		}
	}
	pterm.Printf(
		"obs://%s/%s: size: %d (%s), objects: %d (%d dirs, %d files)\n",
		stats.Bucket, stats.Key,
		stats.Size, internal.FormatSize(stats.Size),
		stats.DirCount+stats.FileCount,
		stats.DirCount, stats.FileCount,
	)
}
